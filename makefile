TARGET=chip8.exe
OUT=build\$(TARGET)
SRC=$(wildcard internal/*.go)
SRC+=main.go

build: $(OUT)

$(OUT): ${SRC}
	go build -gcflags "all=-N -l" -o $(OUT) main.go

clean:
	del /Q $(OUT)