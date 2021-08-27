package main

// import (
// 	"io/ioutil"
// 	"log"
// 	"math/rand"
// 	"os"
// 	"strings"
// 	"time"

// 	"backend.juicedbot.io/juiced.client/http"
// 	_ "backend.juicedbot.io/juiced.client/http/pprof"
// 	rpc "backend.juicedbot.io/juiced.rpc"

// 	api "backend.juicedbot.io/juiced.api"
// 	"backend.juicedbot.io/juiced.infrastructure/common"
// 	"backend.juicedbot.io/juiced.infrastructure/common/captcha"
// 	"backend.juicedbot.io/juiced.infrastructure/common/entities"
// 	"backend.juicedbot.io/juiced.infrastructure/common/enums"
// 	"backend.juicedbot.io/juiced.infrastructure/common/events"
// 	"backend.juicedbot.io/juiced.infrastructure/common/stores"
// 	"backend.juicedbot.io/juiced.infrastructure/queries"
// 	sec "backend.juicedbot.io/juiced.security/auth/util"
// 	"backend.juicedbot.io/juiced.sitescripts/util"

// 	ws "backend.juicedbot.io/juiced.ws"
// 	"github.com/denisbrodbeck/machineid"
// 	"github.com/go-rod/rod/lib/launcher"
// )

// func main() {
// 	go func() {
// 		for {
// 			if os.Getppid() == 1 {
// 				os.Exit(0)
// 			}
// 			time.Sleep(1 * time.Second)
// 		}
// 	}()

// 	hwid, err := machineid.ProtectedID("juiced")
// 	if err != nil {
// 		os.Exit(0)
// 	}

// 	sec.HWID = hwid

// 	go func() {
// 		log.Println(http.ListenAndServe("localhost:5012", nil))
// 	}()

// 	// Initalize the event bus
// 	events.InitEventBus()
// 	eventBus := events.GetEventBus()

// 	// Start the websocket server
// 	go ws.StartWebsocketServer(eventBus)

// 	go func() {
// 		// Wait for the client to connect to the websocket server
// 		channel := make(chan events.Event)
// 		eventBus.Subscribe(channel)
// 		for {
// 			event := <-channel
// 			if event.EventType == events.ConnectEventType {
// 				eventBus.Unsubscribe(channel)
// 				break
// 			}
// 		}

// 		// Initalize the database
// 		err := common.InitDatabase()
// 		if err != nil {
// 			eventBus.PublishCloseEvent()
// 		}

// 		// Get the user's info
// 		_, userInfo, err := queries.GetUserInfo()
// 		if err != nil {
// 			eventBus.PublishCloseEvent()
// 		} else {
// 			enums.UserKey, _, err = sec.GetEncryptionKey(userInfo)
// 			if err != nil {
// 				// No encryption key = no working cards/accounts with saved cards
// 				eventBus.PublishCloseEvent()
// 			} else {
// 				rand.Seed(time.Now().UnixNano())
// 				go Heartbeat(eventBus, userInfo)
// 				go stores.InitTaskStore(eventBus)
// 				stores.InitMonitorStore(eventBus)
// 				stores.InitProxyStore()
// 				captcha.InitCaptchaStore(eventBus)
// 				err := captcha.InitAycd()
// 				if err == nil {
// 					log.Println("Initialized AYCD.")
// 					settings, err := queries.GetSettings()
// 					if err == nil {
// 						if settings.AYCDAccessToken != "" && settings.AYCDAPIKey != "" {
// 							err = captcha.ConnectToAycd(settings.AYCDAccessToken, settings.AYCDAPIKey)
// 							if err != nil {
// 								log.Println("Error connecting to AYCD: " + err.Error())
// 								// TODO @silent: Handle
// 							} else {
// 								log.Println("Connected to AYCD.")
// 							}
// 						}
// 					}
// 				} else {
// 					log.Println("Error initializing AYCD: " + err.Error())
// 					// TODO @silent: Handle
// 				}
// 				go util.DiscordWebhookQueue()
// 				go api.StartServer()

// 				rpc.EnableRPC()
// 				fileInfos, err := ioutil.ReadDir(launcher.DefaultBrowserDir)
// 				if err == nil {
// 					if len(fileInfos) == 0 {
// 						log.Println("Chromium is not installed")
// 						err = launcher.NewBrowser().Download()
// 						if err != nil {
// 							log.Println("Failed to download latest chromium snapshot")
// 						}
// 					}
// 				} else {
// 					log.Println("Failed to find files in default chromium path, trying to download")
// 					err = launcher.NewBrowser().Download()
// 					if err != nil {
// 						log.Println("Failed to download latest chromium snapshot")
// 					}
// 				}
// 				for _, fileInfo := range fileInfos {
// 					if strings.Contains(fileInfo.Name(), "zip") {
// 						if os.Remove(launcher.DefaultBrowserDir+"\\"+fileInfo.Name()) != nil {
// 							log.Println("Could not remove a zip file")
// 						}
// 					}
// 				}

// 			}
// 		}

// 	}()
// 	select {}
// }

// func Heartbeat(eventBus *events.EventBus, userInfo entities.UserInfo) {
// 	lastChecked := time.Now()
// 	var err error
// 	for {
// 		if time.Since(lastChecked).Seconds() > 60 {
// 			userInfo, err = sec.Heartbeat(userInfo, 0)
// 			if err != nil {
// 				eventBus.PublishCloseEvent()
// 			}
// 			lastChecked = time.Now()
// 		}
// 		time.Sleep(common.MS_TO_WAIT)
// 	}
// }
