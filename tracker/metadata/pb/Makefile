PB = $(wildcard *.proto)
GO = $(PB:.proto=.pb.go)

all: $(GO)

%.pb.go: %.proto
		protoc --go_out=plugins=grpc:. $<

clean:
		rm -f *.pb.go
