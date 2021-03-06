#! /usr/bin/env python3
# -*- coding: utf-8 -*-


SERVER_ADDRESS = "192.168.14.12"
SERVER_PORT = 10000

REQUEST_TYPE_PROC_COUNT = 1
REQUEST_TYPE_CONVERT = 2

PLATFORM_IOS = 1
PLATFORM_ANDROID = 2

PLATFORMS_INFO = {
    PLATFORM_IOS: {"paramName": "IOS", "folder": "IOS"},
    PLATFORM_ANDROID: {"paramName": "ANDROID", "folder": "ANDROID"}
}

CONVERT_TYPE_IMAGE_PVR = 1
CONVERT_TYPE_IMAGE_PVRGZ = 2
CONVERT_TYPE_IMAGE_WEBP = 3
CONVERT_TYPE_FFMPEG = 4

PVR_TOOL_PATH = "PVRTexToolCLI"
FFMPEG_TOOL_PATH = "ffmpeg"
WEBP_TOOL_PATH = "cwebp"

NAMES_FILE_NAME = "filesNames.json"

PVR_DEFAULT_PARAMS = "-f PVRTC1_4 -pot + -dither -q pvrtcbest"
PVRGZ16_DEFAULT_PARAMS = "-f r4g4b4a4 -dither"
PVRGZ32_DEFAULT_PARAMS = "-f r8g8b8a8 -dither"
WEBP_DEFAULT_PARAMS = "-q 96"
FFMPEG_DEFAULT_PARAMS = ""

IMAGES_FILES_EXTENTIONS = [".png", ".jpeg", ".jpg"]
SOUNDS_FILES_EXTENTIONS = [".wav", ".ogg", ".mp3"]
VIDEO_FILES_EXTENTIONS = [".mp4", ".mkv", ".avi"]

CONFIG_DATA_SALT = bytes([ 0xBA, 0xBA, 0xEB, 0x53, 0x78, 0x88, 0x32, 0x91 ])

SCRIPT_ROOT_DIR = ""

MAX_PROCESSES_COUNT = 4
REMOTE_SERVER_ACCESSIBLE = False
SERVER_IP_ADDRESS = ""

CONVERT_QUEUE = list()
CONVERT_PROCESSES_LIST = list()
CUSTOM_PARAMS_CONFIG = dict()