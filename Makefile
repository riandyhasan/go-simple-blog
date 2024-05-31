up:
	docker compose up -d --build
stop:
	docker compose stop
down:
	docker compose down

run:
	go run *.go

docker-build:
	docker build -t go-simple-blog .
docker-run:
	docker run -d -p 8080:8080 go-simple-blog