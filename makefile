TARGET=chip8.exe
OUT=build\\$(TARGET)

build: $(OUT)

$(OUT): main.go
	go build -o $(OUT) main.go

clean:
	del /Q $(OUT)