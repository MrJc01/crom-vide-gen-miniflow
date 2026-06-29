.PHONY: build test run clean docker-build docker-run lint

# Variáveis
BIN_NAME=videogen
JSON_TEMPLATE?=templates/sample_v1.json
OUTPUT_FILE?=output_final.mp4

# 60. Makefile para rotinas
build:
	@echo "Compilando $(BIN_NAME)..."
	go build -o $(BIN_NAME) ./cmd/videogen

test:
	@echo "Rodando testes unitários..."
	go test -v -cover ./...

run: build
	@echo "Executando aplicação..."
	./$(BIN_NAME) -json=$(JSON_TEMPLATE) -out=$(OUTPUT_FILE)

# 63. Script de limpeza
clean:
	@echo "Limpando artefatos e pasta temporária..."
	rm -f $(BIN_NAME)
	rm -rf tmp/*.ts tmp/*.mp4 tmp/list.txt
	rm -f $(OUTPUT_FILE)

docker-build:
	@echo "Construindo imagem Docker..."
	docker build -t $(BIN_NAME):latest .

docker-run: docker-build
	@echo "Rodando via Docker..."
	docker run --rm -v $(PWD)/templates:/home/videogen/templates -v $(PWD)/tmp:/home/videogen/tmp -v $(PWD):/home/videogen/output $(BIN_NAME):latest -json=templates/sample_v1.json -out=output/output_final.mp4

lint:
	@echo "Rodando linter..."
	golangci-lint run
