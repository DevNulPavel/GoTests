package game_server

import "log"
import "net/http"
import "golang.org/x/net/websocket"
import _ "net/http/pprof"

type Server struct {
	clients        map[uint32]*Client
	addChannel     chan *Client
	deleteChannel  chan *Client
	sendAllChannel chan bool
}

// Создание нового сервера
func NewServer() *Server {
	clients := make(map[uint32]*Client)
	addChannel := make(chan *Client)
	deleteChannel := make(chan *Client)
	sendAllChannel := make(chan bool)

	return &Server{
		clients,
		addChannel,
		deleteChannel,
		sendAllChannel}
}

// Добавление клиента через очередь
func (server *Server) QueueAddNewClient(c *Client) {
	server.addChannel <- c
}

// Удаление клиента через очередь
func (server *Server) QueueDeleteClient(c *Client) {
	server.deleteChannel <- c
}

// Отправить всем сообщения через очередь
func (server *Server) QueueSendAll() {
	server.sendAllChannel <- true
}

func (server *Server) StartAsyncListen() {
	go server.mainListenFunction()
}

// Отправка всех последних сообщений
func (server *Server) sendStateToClient(c *Client) {
	// Создать состояние текущее
	clientStates := []ClienState{}
	for _, client := range server.clients {
		clientStates = append(clientStates, client.GetState())
	}

	// Отослать юзеру
	c.QueueSendAllStates(clientStates)
}

// Отправить всем сообщение
func (server *Server) sendAllNewState() {
	// Создать состояние текущее
	// Создать состояние текущее
	clientStates := []ClienState{}
	for _, client := range server.clients {
		clientStates = append(clientStates, client.GetState())
	}

	// Отослать всем
	for _, c := range server.clients {
		c.QueueSendAllStates(clientStates)
	}
}

func (server *Server) addClientToMap(client *Client) {
	server.clients[client.id] = client // TODO: TO METHOD
}

func (server *Server) deleteClientFromMap(client *Client) {
	// Даже если нету клиента в мапе - ничего страшного
	delete(server.clients, client.id)
}

func (server *Server) startWebSocketListener() {
	onConnectedHandler := func(ws *websocket.Conn) {
		// Создание нового клиента
		client := NewClient(ws, server)
		server.QueueAddNewClient(client) // выставляем клиента в очередь на добавление (синхронно)
		client.SyncListen()              // блокируется выполнение на данной функции, пока не выйдет клиент

		// Закрытие сокета
		err := ws.Close()
		if err != nil {
			log.Println("Error:", err.Error())
		}

		log.Println("WebSocket connect handler out")
	}
	http.Handle("/websocket", websocket.Handler(onConnectedHandler))
	log.Println("Web socket handler created")
}

// Основная функция прослушивания
func (server *Server) mainListenFunction() {

	log.Println("Listening game_server...")

	// Обработчик подключения WebSocket
	server.startWebSocketListener()

	// Обработка каналов в главной горутине
	for {
		select {
		// Добавление нового юзера
		case c := <-server.addChannel:
			log.Println("Add client")
			server.addClientToMap(c)
			c.QueueSendCurrentClientState() // После добавления на сервере - отправляем клиенту состояние
			server.sendAllNewState()

		// Удаление клиента
		case c := <-server.deleteChannel:
			log.Println("Delete client")
			server.deleteClientFromMap(c)
			server.sendAllNewState()

		// Отправка сообщения всем клиентам
		case <-server.sendAllChannel:
			// Вызываем отправку сообщений всем клиентам
			server.sendAllNewState()
		}
	}
}
