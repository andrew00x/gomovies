.DEFAULT_GOAL=build

CMD_DIR=cmd
SRC_DIRS=$(CMD_DIR) pkg

OUTPUT=gomovies

clean:
	go clean
	rm -f $(OUTPUT)

test: clean
	go test -v $(addprefix ./, $(addsuffix /..., $(SRC_DIRS)))

build: clean
	go build -v -installsuffix "static" -o $(OUTPUT) $(addprefix ./, $(addsuffix /..., $(CMD_DIR)))

install: test
	go install -v -installsuffix "static" $(addprefix ./, $(addsuffix /..., $(CMD_DIR)))
