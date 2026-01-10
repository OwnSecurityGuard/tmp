package main

import (
	"context"
	"fmt"
	"github.com/hashicorp/go-plugin"
	"os/exec"
	"proxy-system-backend/cmd/test/common"
)

func main() {
	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: plugin.HandshakeConfig{
			ProtocolVersion:  1,
			MagicCookieKey:   "GAME_PROTOCOL_PLUGIN",
			MagicCookieValue: "hello",
		},
		Plugins: common.PluginMap,
		Cmd:     exec.Command("E:\\reply\\backend\\cmd\\test\\client\\test.exe"),

		AllowedProtocols: []plugin.Protocol{
			plugin.ProtocolGRPC},
	})
	defer client.Kill()
	rpcClient, err := client.Client()
	if err != nil {
		fmt.Println(err)
		return
	}
	raw, err := rpcClient.Dispense("kv_grpc")

	decoder := raw.(common.GameProtocol)

	info, _ := decoder.Info(context.Background())
	fmt.Println("info ", info.Name)

}
