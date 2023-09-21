# Makefile

APP_NAME = appbff
WATCH_PATH = ./
RELOAD_SCRIPT = run.sh

build:
	go build -o $(APP_NAME)

run: build
	./$(APP_NAME) &

hot-reload: build run watch

watch:
	fswatch -o -e ".*" -i "\\.go$" $(WATCH_PATH) | xargs -n1 -I{} make restart

restart:
	killall -9 $(APP_NAME) || true
	make build run
