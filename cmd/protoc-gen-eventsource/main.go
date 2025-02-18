package main

import (
	"io/ioutil"
	"log"
	"os"

	"github.com/gogo/protobuf/proto"
	plugin_go "github.com/gogo/protobuf/protoc-gen-gogo/plugin"
	"github.com/vectorhacker/eventsource-protobuf/cmd/protoc-gen-eventsource/generate"
)

func check(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

func main() {
	data, err := ioutil.ReadAll(os.Stdin)
	check(err)

	req := plugin_go.CodeGeneratorRequest{}
	err = proto.Unmarshal(data, &req)
	check(err)

	files, err := generate.AllFiles(req.ProtoFile)
	check(err)

	res := &plugin_go.CodeGeneratorResponse{
		File: files,
	}
	data, err = proto.Marshal(res)
	check(err)

	os.Stdout.Write(data)
}
