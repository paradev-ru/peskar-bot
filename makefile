.PHONY: run all linux deploy

run:
	@go run *.go -config-file="./peskar-bot.toml" --log-level="debug"

all:
	@mkdir -p bin/
	@bash --norc -i ./scripts/build.sh

linux:
	@mkdir -p bin/
	@export GOOS=linux && export GOARCH=amd64 && bash --norc -i ./scripts/build.sh

deploy: linux
	@echo "--> Uploading..."
	scp -P 3389 peskar-bot.local leo@paradev.ru:/etc/default/peskar-bot
	scp -P 3389 peskar-bot.toml leo@paradev.ru:/opt/peskar/peskar-bot.toml
	scp -P 3389 contrib/init/sysvinit-debian/peskar-bot leo@paradev.ru:/etc/init.d/peskar-bot
	scp -P 3389 bin/peskar-bot leo@paradev.ru:/opt/peskar/peskar-bot_new
	@echo "--> Restarting..."
	ssh -p 3389 leo@paradev.ru service peskar-bot stop
	ssh -p 3389 leo@paradev.ru rm /opt/peskar/peskar-bot
	ssh -p 3389 leo@paradev.ru mv /opt/peskar/peskar-bot_new /opt/peskar/peskar-bot
	ssh -p 3389 leo@paradev.ru service peskar-bot start
	@echo "--> Getting last logs..."
	@ssh -p 3389 leo@paradev.ru tail -n 25 /var/log/peskar-bot.log
