/// (c) Bernhard Tittelbach, 2019 - MIT License

package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/McKael/madon"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var DebugFlags_ []string
var RegisterAppForTwitterUser_ bool

func init() {
	viper.SetDefault("tag_names", []string{"r3", "realraum"})
	pflag.StringSliceVar(&DebugFlags_, "debug", []string{}, "debug flags e.g. ALL,MADON,MAIN")
	pflag.BoolVar(&RegisterAppForTwitterUser_, "starttwitteroauth", false, "oauth register this app with your twitter user")
	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)
	viper.SetEnvPrefix("MBB")
	viper.AutomaticEnv()
	viper.SetDefault("filterout_sensitive", false)
	viper.SetDefault("filterout_reboosts", true)
	viper.SetDefault("filterfor_accounts_we_follow", true)
}

func main() {
	LogEnable(viper.GetStringSlice("debug")...)

	status_lvl1 := make(chan madon.Notification, 20)
	status_lvl2 := make(chan madon.Notification, 15)

	client, err := madonMustInitClient()
	if err != nil {
		LogMain_.Fatal(err)
	}

	go goSubscribeStreamOfTagNames(client, status_lvl1)
	go goFilterStati(client, status_lvl1, status_lvl2)
	go goFollowBack(client, status_lvl2)

	// wait on Ctrl-C or sigInt or sigKill
	{
		ctrlc_c := make(chan os.Signal, 1)
		signal.Notify(ctrlc_c, os.Interrupt, syscall.SIGTERM)
		<-ctrlc_c //block until ctrl+c is pressed || we receive SIGINT aka kill -1 || kill
		LogMain_.Println("SIGINT received, exiting gracefully ...")
	}
}
