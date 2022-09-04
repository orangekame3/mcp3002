package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/mcp3002"
	"github.com/robfig/cron/v3"
	"github.com/slack-go/slack"
	"golang.org/x/exp/io/spi"
)

func main() {
	dev, err := spi.Open(&spi.Devfs{
		Dev:      "/dev/spidev0.0",
		Mode:     spi.Mode0,
		MaxSpeed: 3600000,
	})
	if err != nil {
		fmt.Println(err)
	}

	defer dev.Close()

	// AD converの設定
	mcp := mcp3002.MCP3002{
		Dev:  dev,
		Vref: 3.3,
	}

	// slackの設定ファイル読み込み
	err = godotenv.Load("../.env")
	if err != nil {
		log.Fatalf("read slack env", err)
	}
	tkn := os.Getenv("TOKEN")
	client := slack.New(tkn)

	var plantA Plant
	plantA = plantA.addADC(mcp)
	plantA = plantA.addChannel(0)

	c := cron.New()
	c.AddFunc("@every 1s", func() {
		if err := worker(plantA, client); err != nil {
			log.Fatal(err)
		}
	})
	c.Start()
	for {
		time.Sleep(time.Second)
	}

}

func worker(p Plant, client *slack.Client) error {
	v, _ := p.ADC.Read()
	p = p.addState(v)
	if p.Status == Tirsty {
		_, _, err := client.PostMessage("#random", slack.MsgOptionText(p.Status, true))
		if err != nil {
			return fmt.Errorf("failed to post message  %w", err)
		}
	}
	return nil
}

const (
	Tirsty = "喉乾いたよ〜"
	Moist  = "お水はもう十分だよ"
)

type Plant struct {
	ADC    mcp3002.MCP3002
	Status string
}

func (p Plant) addState(v float64) Plant {
	resp := p
	if v > 0.5 {
		resp.Status = Tirsty
		return resp
	}
	resp.Status = Moist
	return resp
}

func (p Plant) addChannel(ch int) Plant {
	resp := p
	resp.ADC.Channel = 0
	if ch == 1 {
		resp.ADC.Channel = 1
		return resp
	}
	return resp
}

func (p Plant) addADC(mcp mcp3002.MCP3002) Plant {
	resp := p
	resp.ADC = mcp
	return resp
}
