package gameserver

import (
	"bytes"
	"encoding/binary"
	"io"
	"log"
	"math/rand"
	"net"
	"sync"
	"time"
	"sync/atomic"
)

const UPDATE_QUEUE_SIZE = 100

// Variables
var MAX_ID int32 = 0

// Client ... Структура клиента
type Client struct {
	server            *Server
	connection        *net.TCPConn
	id                int32
	state             ClientState
	mutex             sync.RWMutex
	usersStateChannel chan []ClientState
	exitReadChannel   chan bool
	exitWriteChannel  chan bool
}

// NewClient ... Конструктор
func NewClient(connection *net.TCPConn, server *Server) *Client {
	if connection == nil {
		panic("No connection")
	}
	if server == nil {
		panic("No game server")
	}

	// Увеличиваем id
	curId := atomic.AddInt32(&MAX_ID, 1)

	// Конструируем клиента и его каналы
	clientState := ClientState{curId, rand.Int31() % 100, rand.Int31() % 100, 0}
	usersStateChannel := make(chan []ClientState, UPDATE_QUEUE_SIZE) // В канале апдейтов может накапливаться максимум 1000 апдейтов
	exitReadChannel := make(chan bool, 1)
	exitWriteChannel := make(chan bool, 1)

	return &Client{
		server:            server,
		connection:        connection,
		id:                curId,
		state:             clientState,
		mutex:             sync.RWMutex{},
		usersStateChannel: usersStateChannel,
		exitReadChannel:   exitReadChannel,
		exitWriteChannel:  exitWriteChannel,
	}
}

func (client *Client) Close() {
	(*client.connection).Close()
}

func (client *Client) GetCurrentStateWithTimeReset() ClientState {
	client.mutex.Lock()
	stateCopy := client.state
	client.state.Delta = 0.0
	client.mutex.Unlock()
	return stateCopy
}

// QueueSendAllStates ... Пишем сообщение клиенту
func (client *Client) QueueSendAllStates(states []ClientState) {
	// Если очередь превышена - считаем, что юзер отвалился
	if len(client.usersStateChannel)+1 > UPDATE_QUEUE_SIZE {
		log.Printf("Queue full for client %d", client.id)
		// TODO: Ждем таймаут??
		//client.server.DeleteClient(client)
		//client.exitWriteChannel <- true
		//client.exitReadChannel <- true
		return
	} else {
		client.usersStateChannel <- states
	}
}

// QueueSendCurrentClientState ... Пишем сообщение клиенту только с его состоянием
func (client *Client) QueueSendCurrentClientState() {
	// Если очередь превышена - считаем, что юзер отвалился
	if len(client.usersStateChannel)+1 > UPDATE_QUEUE_SIZE {
		log.Printf("Queue full for client %d", client.id)
		// TODO: Ждем таймаут??
		//client.server.DeleteClient(client)
		//client.exitWriteChannel <- true
		//client.exitReadChannel <- true
		return
	} else {
		client.mutex.RLock()
		currentUserStateCopy := client.state
		client.mutex.RUnlock()

		currentUserStateArray := []ClientState{currentUserStateCopy}

		client.usersStateChannel <- currentUserStateArray
	}
}

// Запускаем ожидания записи и чтения (блокирующая функция)
func (client *Client) StartSyncListenLoop() {
	go client.loopWrite() // в отдельной горутине
	client.loopRead()
}

// Ожидание записи
func (client *Client) loopWrite() {
	//log.Println("StartSyncListenLoop write to client:", client.id)
	for {
		select {
		// Отправка записи клиенту
		case states := <-client.usersStateChannel:
			payloadDataBuffer := new(bytes.Buffer)

			// Количество состояний
			var statesCount uint32 = uint32(len(states))
			binary.Write(payloadDataBuffer, binary.BigEndian, statesCount)

			// Данные
			for _, state := range states {
				stateByteData, err := state.ConvertToBytes()
				if err != nil {
					continue
				}

				payloadDataBuffer.Write(stateByteData)
			}

			// Размер данных
			dataBytes := make([]byte, 4)
			binary.BigEndian.PutUint32(dataBytes, uint32(payloadDataBuffer.Len()))

			// Данные для отправки
			sendData := append(dataBytes, payloadDataBuffer.Bytes()...)

			// Таймаут
			timeout := time.Now().Add(30 * time.Second)
			(*client.connection).SetWriteDeadline(timeout)

			// Отсылаем
			writenCount, err := (*client.connection).Write(sendData)
			if (err != nil) || (writenCount < len(sendData)) {
				client.Close()
				client.server.DeleteClient(client)
				client.exitReadChannel <- true // Выход из loopRead
				if err != nil {
					log.Printf("LoopWrite exit by ERROR (%s), clientId = %d\n", err, client.id)
				} else if writenCount < len(sendData) {
					log.Printf("LoopWrite exit by less bytes - %d from %d, clientId = %d\n", writenCount, len(sendData), client.id)
				}
				return
			}

		// Получение флага выхода из функции
		case <-client.exitWriteChannel:
			log.Println("LoopWrite exit, clientId =", client.id)
			return
		}
	}
}

// Ожидание чтения
func (client *Client) loopRead() {
	//log.Println("Listening read from client")
	for {
		select {
		// Получение флага выхода
		case <-client.exitReadChannel:
			log.Println("LoopRead exit, clientId =", client.id)
			return

		// Чтение данных из сокета
		default:
			// Ожидается, что за 10 минут что-то придет, иначе - это отвал
			timeout := time.Now().Add(10 * time.Minute)
			(*client.connection).SetReadDeadline(timeout)

			// Размер данных
			dataSizeBytes := make([]byte, 4)
			readCount, err := (*client.connection).Read(dataSizeBytes)

			// Ошибка чтения данных
			if (err != nil) || (readCount < 4) {
				client.server.DeleteClient(client)
				client.Close()
				client.exitWriteChannel <- true // для метода loopWrite, чтобы выйти из него

				if err == io.EOF {
					log.Printf("LoopRead exit by disconnect, clientId = %d\n", client.id)
				} else if err != nil {
					log.Printf("LoopRead exit by ERROR (%s), clientId = %d\n", err, client.id)
				} else if readCount < 4 {
					log.Printf("LoopRead exit - read less 8 bytes (%d bytes), clientId = %d\n", readCount, client.id)
				}
				return
			}
			dataSize := binary.BigEndian.Uint32(dataSizeBytes)

			// Ожидается, что будут данные в течении 20 секунд - иначе отвал
			timeout = time.Now().Add(20 * time.Second)
			(*client.connection).SetReadDeadline(timeout)

			// Данные
			data := make([]byte, dataSize)
			readCount, err = (*client.connection).Read(data)

			// Ошибка чтения данных
			if (err != nil) || (uint32(readCount) < dataSize) {
				client.server.DeleteClient(client)
				client.Close()
				client.exitWriteChannel <- true // для метода loopWrite, чтобы выйти из него

				if err == io.EOF {
					log.Printf("LoopRead exit by disconnect, clientId = %d\n", client.id)
				} else if err != nil {
					log.Printf("LoopRead exit by ERROR (%s), clientId = %d\n", err, client.id)
				} else if uint32(readCount) < dataSize {
					log.Printf("LoopRead exit - read less %d bytes (%d bytes), clientId = %d\n", dataSize, readCount, client.id)
				}
				return
			}

			if readCount > 0 {
				// Декодирование
				state, err := NewClientState(data)

				if (err == nil) && (state.ID > 0) {
					// Сбновляем состояние данного клиента
					client.mutex.Lock()
					client.state.X = state.X
					client.state.Y = state.Y
					client.state.Delta += state.Delta
					client.mutex.Unlock()

					// Отправляем обновление состояния всем
					client.server.SendAll()
				}
			}
		}
	}
}
