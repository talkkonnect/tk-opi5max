/*
 * talkkonnect headless mumble client/gateway with lcd screen and channel control
 * Copyright (C) 2018-2019, Suvir Kumar <suvir@talkkonnect.com>
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * Software distributed under the License is distributed on an "AS IS" basis,
 * WITHOUT WARRANTY OF ANY KIND, either express or implied. See the License
 * for the specific language governing rights and limitations under the
 * License.
 *
 * This file has been modified to use the go-gpiocdev library,
 * enabling compatibility with a wider range of Single Board Computers (SBCs)
 * like the Orange Pi, by interacting with the standard Linux GPIO character device.
 *
 * The Initial Developer of the Original Code is
 * Suvir Kumar <suvir@talkkonnect.com>
 * Portions created by the Initial Developer are Copyright (C) Suvir Kumar. All Rights Reserved.
 *
 * Rotary Encoder Alogrithm Inpired By https://www.brainy-bits.com/post/best-code-to-use-with-a-ky-040-rotary-encoder-let-s-find-out
 *
 * Contributor(s):
 *
 * Suvir Kumar <suvir@talkkonnect.com>
 *
 * My Blog is at www.talkkonnect.com
 * The source code is hosted at github.com/talkkonnect
 *
 * gpio.go talkkonnects function to connect to SBC GPIO
 */

package talkkonnect

import (
	"log"
	"strconv"
	"time"

	// [MODIFIED] Replaced rpio and talkkonnect/gpio with the go-gpiocdev library.
	"github.com/warthog618/go-gpiocdev"

	//	"github.com/talkkonnect/go-mcp23017"
	"github.com/talkkonnect/max7219"
)

// [NEW] Global variables for the GPIO chip and line management.
var (
	gpioChip    *gpiocdev.Chip
	outputLines map[string]*gpiocdev.Line
	inputLines  map[string]*gpiocdev.Line
)

// [MODIFIED] Pin variables are now of type *gpiocdev.Line.
var (
	TxButtonUsed  bool
	TxButton      *gpiocdev.Line
	TxButtonPin   uint
	TxButtonState uint

	TxToggleUsed  bool
	TxToggle      *gpiocdev.Line
	TxTogglePin   uint
	TxToggleState uint

	UpButtonUsed  bool
	UpButton      *gpiocdev.Line
	UpButtonPin   uint
	UpButtonState uint

	DownButtonUsed  bool
	DownButton      *gpiocdev.Line
	DownButtonPin   uint
	DownButtonState uint

	PanicUsed        bool
	PanicButton      *gpiocdev.Line
	PanicButtonPin   uint
	PanicButtonState uint

	StreamToggleUsed  bool
	StreamButton      *gpiocdev.Line
	StreamButtonPin   uint
	StreamButtonState uint

	CommentUsed        bool
	CommentButton      *gpiocdev.Line
	CommentButtonPin   uint
	CommentButtonState uint

	ListeningUsed        bool
	ListeningButton      *gpiocdev.Line
	ListeningButtonPin   uint
	ListeningButtonState uint

	RotaryUsed bool
	RotaryA    *gpiocdev.Line
	RotaryB    *gpiocdev.Line
	RotaryAPin uint
	RotaryBPin uint

	RotaryButtonUsed  bool
	RotaryButton      *gpiocdev.Line
	RotaryButtonPin   uint
	RotaryButtonState uint

	VolUpButtonUsed  bool
	VolUpButton      *gpiocdev.Line
	VolUpButtonPin   uint
	VolUpButtonState uint

	VolDownButtonUsed  bool
	VolDownButton      *gpiocdev.Line
	VolDownButtonPin   uint
	VolDownButtonState uint

	TrackingUsed        bool
	TrackingButton      *gpiocdev.Line
	TrackingButtonPin   uint
	TrackingButtonState uint

	MQTT0ButtonUsed  bool
	MQTT0Button      *gpiocdev.Line
	MQTT0ButtonPin   uint
	MQTT0ButtonState uint

	MQTT1ButtonUsed  bool
	MQTT1Button      *gpiocdev.Line
	MQTT1ButtonPin   uint
	MQTT1ButtonState uint

	NextServerButtonUsed  bool
	NextServerButton      *gpiocdev.Line
	NextServerButtonPin   uint
	NextServerButtonState uint

	RepeaterToneButtonUsed  bool
	RepeaterToneButton      *gpiocdev.Line
	RepeaterToneButtonPin   uint
	RepeaterToneButtonState uint

	MemoryChannelButton1Used  bool
	MemoryChannelButton1      *gpiocdev.Line
	MemoryChannelButton1Pin   uint
	MemoryChannelButton1State uint

	MemoryChannelButton2Used  bool
	MemoryChannelButton2      *gpiocdev.Line
	MemoryChannelButton2Pin   uint
	MemoryChannelButton2State uint

	MemoryChannelButton3Used  bool
	MemoryChannelButton3      *gpiocdev.Line
	MemoryChannelButton3Pin   uint
	MemoryChannelButton3State uint

	MemoryChannelButton4Used  bool
	MemoryChannelButton4      *gpiocdev.Line
	MemoryChannelButton4Pin   uint
	MemoryChannelButton4State uint

	ShutdownButtonUsed  bool
	ShutdownButton      *gpiocdev.Line
	ShutdownButtonPin   uint
	ShutdownButtonState uint

	VoiceTargetButton1Used  bool
	VoiceTargetButton1      *gpiocdev.Line
	VoiceTargetButton1Pin   uint
	VoiceTargetButton1State uint

	VoiceTargetButton2Used  bool
	VoiceTargetButton2      *gpiocdev.Line
	VoiceTargetButton2Pin   uint
	VoiceTargetButton2State uint

	VoiceTargetButton3Used  bool
	VoiceTargetButton3      *gpiocdev.Line
	VoiceTargetButton3Pin   uint
	VoiceTargetButton3State uint

	VoiceTargetButton4Used  bool
	VoiceTargetButton4      *gpiocdev.Line
	VoiceTargetButton4Pin   uint
	VoiceTargetButton4State uint

	VoiceTargetButton5Used  bool
	VoiceTargetButton5      *gpiocdev.Line
	VoiceTargetButton5Pin   uint
	VoiceTargetButton5State uint
)

// [NEW] A cleanup function to release GPIO resources gracefully on shutdown.
// This should be called from the main application's cleanup routine (e.g., CleanUp(true)).
func closeGPIO() {
	if gpioChip == nil {
		return
	}
	log.Println("info: Closing GPIO resources...")
	for name, line := range inputLines {
		if err := line.Close(); err != nil {
			log.Printf("warn: could not close input line '%s': %v", name, err)
		}
	}
	for name, line := range outputLines {
		if err := line.Close(); err != nil {
			log.Printf("warn: could not close output line '%s': %v", name, err)
		}
	}
	if err := gpioChip.Close(); err != nil {
		log.Printf("warn: could not close gpiochip: %v", err)
	}
}

// [MODIFIED] The initGPIO function has been significantly rewritten to use go-gpiocdev.
func (b *Talkkonnect) initGPIO() {
	// The check for "rpi" is removed to allow other boards like Orange Pi.
	// GPIO initialization will proceed if enabled in the config.

	// In a real application, the chip name should come from a configuration file.
	// For this conversion, we'll use "gpiochip1" as a sensible default for the Orange Pi 5.
	chipName := "gpiochip1"
	log.Printf("info: Attempting to open GPIO chip '%s'", chipName)

	var err error
	gpioChip, err = gpiocdev.NewChip(chipName)
	if err != nil {
		log.Printf("error: GPIO Error - Failed to open chip '%s'. GPIO is disabled. Error: %v", chipName, err)
		b.GPIOEnabled = false
		return
	}
	b.GPIOEnabled = true

	// Initialize maps to hold the requested GPIO lines.
	outputLines = make(map[string]*gpiocdev.Line)
	inputLines = make(map[string]*gpiocdev.Line)

	// --- Native GPIO Pin Initialization ---
	for _, io := range Config.Global.Hardware.IO.Pins.Pin {
		if io.Enabled && io.Type == "gpio" {
			if io.Direction == "input" {
				line, err := gpioChip.RequestLine(int(io.PinNo), gpiocdev.AsInput, gpiocdev.WithPullUp)
				if err != nil {
					log.Printf("error: Failed to request GPIO input line %d (%s): %v", io.PinNo, io.Name, err)
					continue
				}
				log.Printf("debug: GPIO Setup Input Device %v Name %v PinNo %v", io.Device, io.Name, io.PinNo)
				inputLines[io.Name] = line

				// Map the generic line to a specific function variable.
				if io.Name == "txptt" {
					TxButton, TxButtonUsed, TxButtonPin = line, true, io.PinNo
				} else if io.Name == "txtoggle" {
					TxToggle, TxToggleUsed, TxTogglePin = line, true, io.PinNo
				} else if io.Name == "channelup" {
					UpButton, UpButtonUsed, UpButtonPin = line, true, io.PinNo
				} else if io.Name == "channeldown" {
					DownButton, DownButtonUsed, DownButtonPin = line, true, io.PinNo
				} else if io.Name == "panic" {
					PanicButton, PanicUsed, PanicButtonPin = line, true, io.PinNo
				} else if io.Name == "streamtoggle" {
					StreamButton, StreamToggleUsed, StreamButtonPin = line, true, io.PinNo
				} else if io.Name == "comment" {
					CommentButton, CommentUsed, CommentButtonPin = line, true, io.PinNo
				} else if io.Name == "listening" {
					ListeningButton, ListeningUsed, ListeningButtonPin = line, true, io.PinNo
				} else if io.Name == "rotarya" {
					RotaryA, RotaryUsed, RotaryAPin = line, true, io.PinNo
				} else if io.Name == "rotaryb" {
					RotaryB, RotaryUsed, RotaryBPin = line, true, io.PinNo
				} else if io.Name == "rotarybutton" {
					RotaryButton, RotaryButtonUsed, RotaryButtonPin = line, true, io.PinNo
				} else if io.Name == "volup" {
					VolUpButton, VolUpButtonUsed, VolUpButtonPin = line, true, io.PinNo
				} else if io.Name == "voldown" {
					VolDownButton, VolDownButtonUsed, VolDownButtonPin = line, true, io.PinNo
				} else if io.Name == "tracking" {
					TrackingButton, TrackingUsed, TrackingButtonPin = line, true, io.PinNo
				} else if io.Name == "mqtt0" {
					MQTT0Button, MQTT0ButtonUsed, MQTT0ButtonPin = line, true, io.PinNo
				} else if io.Name == "mqtt1" {
					MQTT1Button, MQTT1ButtonUsed, MQTT1ButtonPin = line, true, io.PinNo
				} else if io.Name == "nextserver" {
					NextServerButton, NextServerButtonUsed, NextServerButtonPin = line, true, io.PinNo
				} else if io.Name == "memorychannel1" {
					MemoryChannelButton1, MemoryChannelButton1Used, MemoryChannelButton1Pin = line, true, io.PinNo
				} else if io.Name == "memorychannel2" {
					MemoryChannelButton2, MemoryChannelButton2Used, MemoryChannelButton2Pin = line, true, io.PinNo
				} else if io.Name == "memorychannel3" {
					MemoryChannelButton3, MemoryChannelButton3Used, MemoryChannelButton3Pin = line, true, io.PinNo
				} else if io.Name == "memorychannel4" {
					MemoryChannelButton4, MemoryChannelButton4Used, MemoryChannelButton4Pin = line, true, io.PinNo
				} else if io.Name == "repeatertone" {
					RepeaterToneButton, RepeaterToneButtonUsed, RepeaterToneButtonPin = line, true, io.PinNo
				} else if io.Name == "shutdown" {
					ShutdownButton, ShutdownButtonUsed, ShutdownButtonPin = line, true, io.PinNo
				} else if io.Name == "presetvoicetarget1" {
					VoiceTargetButton1, VoiceTargetButton1Used, VoiceTargetButton1Pin = line, true, io.PinNo
				} else if io.Name == "presetvoicetarget2" {
					VoiceTargetButton2, VoiceTargetButton2Used, VoiceTargetButton2Pin = line, true, io.PinNo
				} else if io.Name == "presetvoicetarget3" {
					VoiceTargetButton3, VoiceTargetButton3Used, VoiceTargetButton3Pin = line, true, io.PinNo
				} else if io.Name == "presetvoicetarget4" {
					VoiceTargetButton4, VoiceTargetButton4Used, VoiceTargetButton4Pin = line, true, io.PinNo
				} else if io.Name == "presetvoicetarget5" {
					VoiceTargetButton5, VoiceTargetButton5Used, VoiceTargetButton5Pin = line, true, io.PinNo
				}
			} else if io.Direction == "output" {
				line, err := gpioChip.RequestLine(int(io.PinNo), gpiocdev.AsInput, gpiocdev.WithPullUp)
				if err != nil {
					log.Printf("error: Failed to request GPIO output line %d (%s): %v", io.PinNo, io.Name, err)
					continue
				}
				line.Reconfigure(gpiocdev.AsOutput(int(io.PinNo)))
				log.Printf("debug: GPIO Setup Output Device %v Name %v PinNo %v", io.Device, io.Name, io.PinNo)
				outputLines[io.Name] = line
			}
		}
	}

	// [MODIFIED] Goroutines now use line.Value() instead of pin.Read()
	if TxButtonUsed {
		go func() {
			for {
				if IsConnected && TxButtonUsed {
					time.Sleep(150 * time.Millisecond)
					val, err := TxButton.Value()
					if err != nil {
						log.Printf("error: reading TxButton: %v", err)
						continue
					}
					currentState := uint(val)
					if currentState != TxButtonState {
						TxButtonState = currentState
						if b.Stream != nil {
							if TxButtonState == 1 { // Released (High for pull-up)
								if isTx {
									isTx = false
									b.TransmitStop(true)
									playIOMedia("iotxpttstop")
									if Config.Global.Software.Settings.TxCounter {
										txcounter++
										log.Println("debug: Tx Button Count ", txcounter)
									}
								}
							} else { // Pressed (Low for pull-up)
								log.Println("debug: Tx Button is pressed")
								if !isTx {
									isTx = true
									playIOMedia("iotxpttstart")
								}
								txlockout := &TXLockOut
								if Config.Global.Software.Settings.TXLockOut && *txlockout {
									log.Println("warn: TX Lockout Stopping Transmission")
									eventSound := findEventSound("txlockout")
									if eventSound.Enabled {
										if v, err := strconv.Atoi(eventSound.Volume); err == nil {
											localMediaPlayer(eventSound.FileName, v, eventSound.Blocking, 0, 1)
											log.Printf("debug: Playing txlockout Sound")
										}
									}
								} else {
									b.TransmitStart()
								}
							}
						}
					}
				} else {
					time.Sleep(1 * time.Second)
				}
			}
		}()
	}

	if TxToggleUsed {
		go func() {
			var prevState uint = 1
			for {
				if IsConnected && TxToggleUsed {
					val, err := TxToggle.Value()
					if err != nil {
						log.Println("error: Error reading TXToggle Pin:", err)
						time.Sleep(1 * time.Second)
						continue
					}
					currentState := uint(val)
					time.Sleep(150 * time.Millisecond)

					if currentState != prevState && currentState == 0 { // Action on press (going to low)
						isTx = !isTx
						if !isTx { // Was transmitting, now stop
							b.TransmitStop(true)
							log.Println("debug: Toggle Stopped Transmitting")
							playIOMedia("iotxtogglestop")
						} else { // Was not transmitting, now start
							playIOMedia("txtogglestart")
							b.TransmitStart()
							log.Println("debug: Toggle Started Transmitting")
						}
					}
					prevState = currentState
				} else {
					time.Sleep(1 * time.Second)
				}
			}
		}()
	}

	if UpButtonUsed {
		go func() {
			for {
				if IsConnected && UpButtonUsed {
					val, err := UpButton.Value()
					if err != nil {
						log.Printf("error: reading UpButton: %v", err)
						time.Sleep(1 * time.Second)
						continue
					}
					currentState := uint(val)
					if currentState != UpButtonState {
						UpButtonState = currentState
						if UpButtonState == 1 {
							log.Println("debug: UP Button is released")
						} else {
							log.Println("debug: UP Button is pressed")
							playIOMedia("iochannelup")
							b.ChannelUp()
						}
					}
					time.Sleep(150 * time.Millisecond)
				} else {
					time.Sleep(1 * time.Second)
				}
			}
		}()
	}

	if DownButtonUsed {
		go func() {
			for {
				if IsConnected && DownButtonUsed {
					val, err := DownButton.Value()
					if err != nil {
						log.Printf("error: reading DownButton: %v", err)
						time.Sleep(1 * time.Second)
						continue
					}
					currentState := uint(val)
					if currentState != DownButtonState {
						DownButtonState = currentState
						if DownButtonState == 1 {
							log.Println("debug: Ch Down Button is released")
						} else {
							log.Println("debug: Ch Down Button is pressed")
							playIOMedia("iochanneldown")
							b.ChannelDown()
						}
					}
					time.Sleep(150 * time.Millisecond)
				} else {
					time.Sleep(1 * time.Second)
				}
			}
		}()
	}

	if PanicUsed {
		go func() {
			for {
				if IsConnected && PanicUsed {
					val, err := PanicButton.Value()
					if err != nil {
						log.Printf("error: reading PanicButton: %v", err)
						time.Sleep(1 * time.Second)
						continue
					}
					currentState := uint(val)
					if currentState != PanicButtonState {
						PanicButtonState = currentState
						if PanicButtonState == 1 {
							log.Println("debug: Panic Button is released")
						} else {
							log.Println("debug: Panic Button is pressed")
							playIOMedia("iopanic")
							b.cmdPanicSimulation()
						}
					}
					time.Sleep(150 * time.Millisecond)
				} else {
					time.Sleep(1 * time.Second)
				}
			}
		}()
	}

	if CommentUsed {
		go func() {
			for {
				if IsConnected && CommentUsed {
					val, err := CommentButton.Value()
					if err != nil {
						log.Printf("error: reading CommentButton: %v", err)
						time.Sleep(1 * time.Second)
						continue
					}
					currentState := uint(val)
					if currentState != CommentButtonState {
						CommentButtonState = currentState
						if CommentButtonState == 1 {
							playIOMedia("iocommenton")
							log.Println("debug: Comment Button State 1 setting comment to State 1 Message ", Config.Global.Hardware.Comment.CommentMessageOff)
							b.SetComment(Config.Global.Hardware.Comment.CommentMessageOff)
						} else {
							playIOMedia("iocommentoff")
							log.Println("debug: Comment Button State 2 setting comment to State 2 Message ", Config.Global.Hardware.Comment.CommentMessageOn)
							b.SetComment(Config.Global.Hardware.Comment.CommentMessageOn)
						}
					}
					time.Sleep(150 * time.Millisecond)
				} else {
					time.Sleep(1 * time.Second)
				}
			}
		}()
	}

	if ListeningUsed {
		go func() {
			for {
				if IsConnected && ListeningUsed {
					val, err := ListeningButton.Value()
					if err != nil {
						log.Printf("error: reading ListeningButton: %v", err)
						time.Sleep(1 * time.Second)
						continue
					}
					currentState := uint(val)
					if currentState != ListeningButtonState {
						ListeningButtonState = currentState
						if ListeningButtonState == 1 {
							playIOMedia("iolisteningstop")
							log.Println("debug: Listening Button State 1 Listening Stop")
							b.listeningToChannels("stop")
						} else {
							playIOMedia("iolisteningstart")
							b.listeningToChannels("start")
							log.Println("debug: Listening Button State 0 Listening Start")
						}
					}
					time.Sleep(150 * time.Millisecond)
				} else {
					time.Sleep(1 * time.Second)
				}
			}
		}()
	}

	if StreamToggleUsed {
		go func() {
			for {
				if IsConnected && StreamToggleUsed {
					val, err := StreamButton.Value()
					if err != nil {
						log.Printf("error: reading StreamButton: %v", err)
						time.Sleep(1 * time.Second)
						continue
					}
					currentState := uint(val)
					if currentState != StreamButtonState {
						StreamButtonState = currentState
						if StreamButtonState == 1 {
							log.Println("debug: Stream Button is released")
						} else {
							playIOMedia("iostreamtoggle")
							log.Println("debug: Stream Button is pressed")
							b.cmdPlayback()
						}
					}
					time.Sleep(150 * time.Millisecond)
				} else {
					time.Sleep(1 * time.Second)
				}
			}
		}()
	}

	if RotaryUsed {
		go func() {
			var valA, valB, lastValA, lastValB int
			for {
				if IsConnected && RotaryUsed {
					valA, _ = RotaryA.Value()
					valB, _ = RotaryB.Value()
					time.Sleep(2 * time.Millisecond)
					lastValA, _ = RotaryA.Value()
					lastValB, _ = RotaryB.Value()

					if lastValA == 0 && lastValB == 1 {
						if valA == 1 && valB == 0 {
							b.rotaryAction("ccw")
							continue
						}
						if valA == 1 && valB == 1 {
							b.rotaryAction("cw")
							continue
						}
					}

					if lastValA == 1 && lastValB == 0 {
						if valA == 0 && valB == 1 {
							b.rotaryAction("ccw")
							continue
						}
						if valA == 0 && valB == 0 {
							b.rotaryAction("cw")
							continue
						}
					}

					if lastValA == 1 && lastValB == 1 {
						if valA == 0 && valB == 1 {
							b.rotaryAction("ccw")
							continue
						}
						if valA == 0 && valB == 0 {
							b.rotaryAction("cw")
							continue
						}
					}

					if lastValA == 0 && lastValB == 0 {
						if valA == 1 && valB == 0 {
							b.rotaryAction("ccw")
							continue
						}
						if valA == 1 && valB == 1 {
							b.rotaryAction("cw")
							continue
						}
					}
				} else {
					time.Sleep(1 * time.Second)
				}
			}
		}()
	}

	if RotaryButtonUsed {
		go func() {
			for {
				if IsConnected && RotaryButtonUsed {
					val, err := RotaryButton.Value()
					if err != nil {
						log.Printf("error: reading RotaryButton: %v", err)
						time.Sleep(1 * time.Second)
						continue
					}
					currentState := uint(val)
					if currentState != RotaryButtonState {
						RotaryButtonState = currentState
						if RotaryButtonState == 1 {
							log.Println("debug: Rotary Button is released")
						} else {
							log.Println("debug: Rotary Button is pressed")
							playIOMedia("iorotarybutton")
							b.nextEnabledRotaryEncoderFunction()
						}
					}
					time.Sleep(150 * time.Millisecond)
				} else {
					time.Sleep(1 * time.Second)
				}
			}
		}()
	}

	if VolUpButtonUsed {
		go func() {
			for {
				if IsConnected && VolUpButtonUsed {
					val, err := VolUpButton.Value()
					if err != nil {
						log.Printf("error: reading VolUpButton: %v", err)
						time.Sleep(1 * time.Second)
						continue
					}
					currentState := uint(val)
					if currentState != VolUpButtonState {
						VolUpButtonState = currentState
						if VolUpButtonState == 1 {
							log.Println("debug: Vol UP Button is released")
						} else {
							log.Println("debug: Vol UP Button is pressed")
							playIOMedia("iovolup")
							b.cmdVolumeRXUp()
						}
					}
					time.Sleep(150 * time.Millisecond)
				} else {
					time.Sleep(1 * time.Second)
				}
			}
		}()
	}

	if VolDownButtonUsed {
		go func() {
			for {
				if IsConnected && VolDownButtonUsed {
					val, err := VolDownButton.Value()
					if err != nil {
						log.Printf("error: reading VolDownButton: %v", err)
						time.Sleep(1 * time.Second)
						continue
					}
					currentState := uint(val)
					if currentState != VolDownButtonState {
						VolDownButtonState = currentState
						if VolDownButtonState == 1 {
							log.Println("debug: Vol Down Button is released")
						} else {
							log.Println("debug: Vol Down Button is pressed")
							playIOMedia("iovoldown")
							b.cmdVolumeRXDown()
						}
					}
					time.Sleep(150 * time.Millisecond)
				} else {
					time.Sleep(1 * time.Second)
				}
			}
		}()
	}

	if TrackingUsed {
		go func() {
			for {
				if IsConnected && TrackingUsed {
					val, err := TrackingButton.Value()
					if err != nil {
						log.Printf("error: reading TrackingButton: %v", err)
						time.Sleep(1 * time.Second)
						continue
					}
					currentState := uint(val)
					if currentState != TrackingButtonState {
						TrackingButtonState = currentState
						if TrackingButtonState == 1 {
							playIOMedia("iotrackingon")
							log.Println("debug: Tracking Button State 1 setting GPS Tracking on  ")
							// place holder to start tracking timer
						} else {
							playIOMedia("iotrackingoff")
							log.Println("debug: Tracking Button State 1 setting GPS Tracking off ")
							// place holder to start tracking timer
						}
					}
					time.Sleep(150 * time.Millisecond)
				} else {
					time.Sleep(1 * time.Second)
				}
			}
		}()
	}

	if MQTT0ButtonUsed {
		go func() {
			for {
				if IsConnected && MQTT0ButtonUsed {
					val, err := MQTT0Button.Value()
					if err != nil {
						log.Printf("error: reading MQTT0Button: %v", err)
						time.Sleep(1 * time.Second)
						continue
					}
					currentState := uint(val)
					if currentState != MQTT0ButtonState {
						MQTT0ButtonState = currentState
						if MQTT0ButtonState == 1 {
							log.Println("debug: MQTT0 Button is released")
						} else {
							log.Println("debug: MQTT0 Button is pressed")
							playIOMedia("iomqtt0")
							MQTTButtonCommand := findMQTTButton("0")
							if MQTTButtonCommand.Enabled {
								MQTTPublish(MQTTButtonCommand.Payload)
							}
						}
					}
					time.Sleep(150 * time.Millisecond)
				} else {
					time.Sleep(1 * time.Second)
				}
			}
		}()
	}

	if MQTT1ButtonUsed {
		go func() {
			for {
				if IsConnected && MQTT1ButtonUsed {
					val, err := MQTT1Button.Value()
					if err != nil {
						log.Printf("error: reading MQTT1Button: %v", err)
						time.Sleep(1 * time.Second)
						continue
					}
					currentState := uint(val)
					if currentState != MQTT1ButtonState {
						MQTT1ButtonState = currentState
						if MQTT1ButtonState == 1 {
							log.Println("debug: MQTT1 Button is released")
						} else {
							log.Println("debug: MQTT1 Button is pressed")
							playIOMedia("iomqtt1")
							MQTTButtonCommand := findMQTTButton("1")
							if MQTTButtonCommand.Enabled {
								MQTTPublish(MQTTButtonCommand.Payload)
							}
						}
					}
					time.Sleep(150 * time.Millisecond)
				} else {
					time.Sleep(1 * time.Second)
				}
			}
		}()
	}

	if NextServerButtonUsed {
		go func() {
			for {
				if IsConnected && NextServerButtonUsed {
					val, err := NextServerButton.Value()
					if err != nil {
						log.Printf("error: reading NextServerButton: %v", err)
						time.Sleep(1 * time.Second)
						continue
					}
					currentState := uint(val)
					if currentState != NextServerButtonState {
						NextServerButtonState = currentState
						if NextServerButtonState == 1 {
							log.Println("debug: NextServer Button is released")
						} else {
							log.Println("debug: NextServer Button is pressed")
							playIOMedia("iocnextserver")
							b.cmdConnNextServer()
						}
					}
					time.Sleep(150 * time.Millisecond)
				} else {
					time.Sleep(1 * time.Second)
				}
			}
		}()
	}

	if MemoryChannelButton1Used {
		go func() {
			for {
				if IsConnected && MemoryChannelButton1Used {
					val, err := MemoryChannelButton1.Value()
					if err != nil {
						log.Printf("error: reading MemoryChannelButton1: %v", err)
						time.Sleep(1 * time.Second)
						continue
					}
					currentState := uint(val)
					if currentState != MemoryChannelButton1State {
						MemoryChannelButton1State = currentState
						if MemoryChannelButton1State == 1 {
							log.Println("debug: MemoryChannelButton1 Button is released")
						} else {
							log.Println("debug: MemoryChannelButton1 Button is pressed")
							playIOMedia("memorychannel")
							v, found := GPIOMemoryMap["memorychannel1"]
							if found {
								b.ChangeChannel(v.ChannelName)
							} else {
								log.Printf("error: Channel %v Not Found Channel Change Failed\n", v)
							}
						}
					}
					time.Sleep(150 * time.Millisecond)
				} else {
					time.Sleep(1 * time.Second)
				}
			}
		}()
	}

	if MemoryChannelButton2Used {
		go func() {
			for {
				if IsConnected && MemoryChannelButton2Used {
					val, err := MemoryChannelButton2.Value()
					if err != nil {
						log.Printf("error: reading MemoryChannelButton2: %v", err)
						time.Sleep(1 * time.Second)
						continue
					}
					currentState := uint(val)
					if currentState != MemoryChannelButton2State {
						MemoryChannelButton2State = currentState
						if MemoryChannelButton2State == 1 {
							log.Println("debug: MemoryChannelButton2 Button is released")
						} else {
							log.Println("debug: MemoryChannelButton2 Button is pressed")
							playIOMedia("memorychannel")
							v, found := GPIOMemoryMap["memorychannel2"]
							if found {
								b.ChangeChannel(v.ChannelName)
							} else {
								log.Printf("error: Channel %v Not Found Channel Change Failed\n", v)
							}
						}
					}
					time.Sleep(150 * time.Millisecond)
				} else {
					time.Sleep(1 * time.Second)
				}
			}
		}()
	}

	if MemoryChannelButton3Used {
		go func() {
			for {
				if IsConnected && MemoryChannelButton3Used {
					val, err := MemoryChannelButton3.Value()
					if err != nil {
						log.Printf("error: reading MemoryChannelButton3: %v", err)
						time.Sleep(1 * time.Second)
						continue
					}
					currentState := uint(val)
					if currentState != MemoryChannelButton3State {
						MemoryChannelButton3State = currentState
						if MemoryChannelButton3State == 1 {
							log.Println("debug: MemoryChannelButton3 Button is released")
						} else {
							log.Println("debug: MemoryChannelButton3 Button is pressed")
							playIOMedia("memorychannel")
							v, found := GPIOMemoryMap["memorychannel3"]
							if found {
								b.ChangeChannel(v.ChannelName)
							} else {
								log.Printf("error: Channel %v Not Found Channel Change Failed\n", v)
							}
						}
					}
					time.Sleep(150 * time.Millisecond)
				} else {
					time.Sleep(1 * time.Second)
				}
			}
		}()
	}

	if MemoryChannelButton4Used {
		go func() {
			for {
				if IsConnected && MemoryChannelButton4Used {
					val, err := MemoryChannelButton4.Value()
					if err != nil {
						log.Printf("error: reading MemoryChannelButton4: %v", err)
						time.Sleep(1 * time.Second)
						continue
					}
					currentState := uint(val)
					if currentState != MemoryChannelButton4State {
						MemoryChannelButton4State = currentState
						if MemoryChannelButton4State == 1 {
							log.Println("debug: MemoryChannelButton4 Button is released")
						} else {
							log.Println("debug: MemoryChannelButton4 Button is pressed")
							playIOMedia("memorychannel")
							v, found := GPIOMemoryMap["memorychannel4"]
							if found {
								b.ChangeChannel(v.ChannelName)
							} else {
								log.Printf("error: Channel %v Not Found Channel Change Failed\n", v)
							}
						}
					}
					time.Sleep(150 * time.Millisecond)
				} else {
					time.Sleep(1 * time.Second)
				}
			}
		}()
	}

	if VoiceTargetButton1Used {
		go func() {
			for {
				if IsConnected && VoiceTargetButton1Used {
					val, err := VoiceTargetButton1.Value()
					if err != nil {
						log.Printf("error: reading VoiceTargetButton1: %v", err)
						time.Sleep(1 * time.Second)
						continue
					}
					currentState := uint(val)
					if currentState != VoiceTargetButton1State {
						VoiceTargetButton1State = currentState
						if VoiceTargetButton1State == 1 {
							log.Println("debug: VoiceTargetButton1 Button is released")
						} else {
							log.Println("debug: VoicetargetButton1 Button is pressed")
							playIOMedia("voicetarget1")
							vtid, found := GPIOVoiceTargetMap["presetvoicetarget1"]
							if found {
								b.cmdSendVoiceTargets(vtid.ID)
								log.Printf("info: Setting Voice Target to ID %v\n", vtid.ID)
							} else {
								log.Printf("error: VoicetargetButton1 Mapped to ID %v Not Found VoiceTarget Set Failed\n", vtid.ID)
							}
						}
					}
					time.Sleep(150 * time.Millisecond)
				} else {
					time.Sleep(1 * time.Second)
				}
			}
		}()
	}

	if VoiceTargetButton2Used {
		go func() {
			for {
				if IsConnected && VoiceTargetButton2Used {
					val, err := VoiceTargetButton2.Value()
					if err != nil {
						log.Printf("error: reading VoiceTargetButton2: %v", err)
						time.Sleep(1 * time.Second)
						continue
					}
					currentState := uint(val)
					if currentState != VoiceTargetButton2State {
						VoiceTargetButton2State = currentState
						if VoiceTargetButton2State == 1 {
							log.Println("debug: VoiceTargetButton2 Button is released")
						} else {
							log.Println("debug: VoicetargetButton2 Button is pressed")
							playIOMedia("voicetarget2")
							vtid, found := GPIOVoiceTargetMap["presetvoicetarget2"]
							if found {
								b.cmdSendVoiceTargets(vtid.ID)
								log.Printf("info: Setting Voice Target to ID %v\n", vtid.ID)
							} else {
								log.Printf("error: VoicetargetButton2 Mapped to ID %v Not Found VoiceTarget Set Failed\n", vtid.ID)
							}
						}
					}
					time.Sleep(150 * time.Millisecond)
				} else {
					time.Sleep(1 * time.Second)
				}
			}
		}()
	}

	if VoiceTargetButton3Used {
		go func() {
			for {
				if IsConnected && VoiceTargetButton3Used {
					val, err := VoiceTargetButton3.Value()
					if err != nil {
						log.Printf("error: reading VoiceTargetButton3: %v", err)
						time.Sleep(1 * time.Second)
						continue
					}
					currentState := uint(val)
					if currentState != VoiceTargetButton3State {
						VoiceTargetButton3State = currentState
						if VoiceTargetButton3State == 1 {
							log.Println("debug: VoiceTargetButton3 Button is released")
						} else {
							log.Println("debug: VoicetargetButton3 Button is pressed")
							playIOMedia("voicetarget3")
							vtid, found := GPIOVoiceTargetMap["presetvoicetarget3"]
							if found {
								b.cmdSendVoiceTargets(vtid.ID)
								log.Printf("info: Setting Voice Target to ID %v\n", vtid.ID)
							} else {
								log.Printf("error: VoicetargetButton3 Mapped to ID %v Not Found VoiceTarget Set Failed\n", vtid.ID)
							}
						}
					}
					time.Sleep(150 * time.Millisecond)
				} else {
					time.Sleep(1 * time.Second)
				}
			}
		}()
	}

	if VoiceTargetButton4Used {
		go func() {
			for {
				if IsConnected && VoiceTargetButton4Used {
					val, err := VoiceTargetButton4.Value()
					if err != nil {
						log.Printf("error: reading VoiceTargetButton4: %v", err)
						time.Sleep(1 * time.Second)
						continue
					}
					currentState := uint(val)
					if currentState != VoiceTargetButton4State {
						VoiceTargetButton4State = currentState
						if VoiceTargetButton4State == 1 {
							log.Println("debug: VoiceTargetButton4 Button is released")
						} else {
							log.Println("debug: VoicetargetButton4 Button is pressed")
							playIOMedia("voicetarget4")
							vtid, found := GPIOVoiceTargetMap["presetvoicetarget4"]
							if found {
								b.cmdSendVoiceTargets(vtid.ID)
								log.Printf("info: Setting Voice Target to ID %v\n", vtid.ID)
							} else {
								log.Printf("error: VoicetargetButton4 Mapped to ID %v Not Found VoiceTarget Set Failed\n", vtid.ID)
							}
						}
					}
					time.Sleep(150 * time.Millisecond)
				} else {
					time.Sleep(1 * time.Second)
				}
			}
		}()
	}

	if VoiceTargetButton5Used {
		go func() {
			for {
				if IsConnected && VoiceTargetButton5Used {
					val, err := VoiceTargetButton5.Value()
					if err != nil {
						log.Printf("error: reading VoiceTargetButton5: %v", err)
						time.Sleep(1 * time.Second)
						continue
					}
					currentState := uint(val)
					if currentState != VoiceTargetButton5State {
						VoiceTargetButton5State = currentState
						if VoiceTargetButton5State == 1 {
							log.Println("debug: VoiceTargetButton5 Button is released")
						} else {
							log.Println("debug: VoicetargetButton5 Button is pressed")
							playIOMedia("voicetarget5")
							vtid, found := GPIOVoiceTargetMap["presetvoicetarget5"]
							if found {
								b.cmdSendVoiceTargets(vtid.ID)
								log.Printf("info: Setting Voice Target to ID %v\n", vtid.ID)
							} else {
								log.Printf("error: VoicetargetButton5 Mapped to ID %v Not Found VoiceTarget Set Failed\n", vtid.ID)
							}
						}
					}
					time.Sleep(150 * time.Millisecond)
				} else {
					time.Sleep(1 * time.Second)
				}
			}
		}()
	}

	if RepeaterToneButtonUsed {
		go func() {
			for {
				if IsConnected && RepeaterToneButtonUsed {
					val, err := RepeaterToneButton.Value()
					if err != nil {
						log.Printf("error: reading RepeaterToneButton: %v", err)
						time.Sleep(1 * time.Second)
						continue
					}
					currentState := uint(val)
					if currentState != RepeaterToneButtonState {
						RepeaterToneButtonState = currentState
						if RepeaterToneButtonState == 1 {
							log.Println("debug: Repeater Tone Button is released")
						} else {
							log.Println("debug: Repeater Tone Button is pressed")
							playIOMedia("iorepeatertone")
							b.cmdPlayRepeaterTone()
						}
					}
					time.Sleep(150 * time.Millisecond)
				} else {
					time.Sleep(1 * time.Second)
				}
			}
		}()
	}

	if ShutdownButtonUsed {
		go func() {
			for {
				if IsConnected && ShutdownButtonUsed {
					val, err := ShutdownButton.Value()
					if err != nil {
						log.Printf("error: reading ShutdownButton: %v", err)
						time.Sleep(1 * time.Second)
						continue
					}
					currentState := uint(val)
					if currentState != ShutdownButtonState {
						ShutdownButtonState = currentState
						if ShutdownButtonState == 1 {
							log.Println("debug: Shutdown is released")
						} else {
							log.Println("debug: Shutdown Button is pressed")
							playIOMedia("shutdown")
							duration := time.Since(StartTime)
							log.Printf("info: Talkkonnect Now Running For %v \n", secondsToHuman(int(duration.Seconds())))
							b.sevenSegment("bye", "")
							TTSEvent("quittalkkonnect")
							closeGPIO() // Gracefully release GPIO resources
							CleanUp(true)
						}
					}
					time.Sleep(150 * time.Millisecond)
				} else {
					time.Sleep(1 * time.Second)
				}
			}
		}()
	}
}

// [MODIFIED] GPIOOutPin is updated to use the pre-initialized lines from the outputLines map.
func GPIOOutPin(name string, command string) {
	for _, io := range Config.Global.Hardware.IO.Pins.Pin {
		if io.Enabled && io.Direction == "output" && io.Name == name {
			switch io.Type {
			case "gpio":
				line, ok := outputLines[name]
				if !ok {
					log.Printf("error: GPIO line '%s' (pin %d) was not initialized.", name, io.PinNo)
					return
				}

				var val int
				if command == "on" {
					val = 1
					if io.Inverted {
						val = 0
					}
					log.Printf("debug: Turning On %v at pin %v Output GPIO (Inverted: %v)\n", io.Name, io.PinNo, io.Inverted)
					if err := line.SetValue(val); err != nil {
						log.Printf("error: Failed to set value for line %s: %v", name, err)
					}
				} else if command == "off" {
					val = 0
					if io.Inverted {
						val = 1
					}
					log.Printf("debug: Turning Off %v at pin %v Output GPIO (Inverted: %v)\n", io.Name, io.PinNo, io.Inverted)
					if err := line.SetValue(val); err != nil {
						log.Printf("error: Failed to set value for line %s: %v", name, err)
					}
				} else if command == "pulse" {
					log.Printf("debug: Pulsing %v at pin %v Output GPIO\n", io.Name, io.PinNo)
					onVal, offVal := 1, 0
					if io.Inverted {
						onVal, offVal = 0, 1
					}
					_ = line.SetValue(offVal)
					time.Sleep(Config.Global.Hardware.IO.Pulse.Leading * time.Millisecond)
					_ = line.SetValue(onVal)
					time.Sleep(Config.Global.Hardware.IO.Pulse.Pulse * time.Millisecond)
					_ = line.SetValue(offVal)
					time.Sleep(Config.Global.Hardware.IO.Pulse.Trailing * time.Millisecond)
				}
			default:
				log.Println("error: GPIO Types Currently Supported are gpio or mcp23017 only!")
			}
			break // Pin found and command handled
		}
	}
}

// [MODIFIED] GPIOOutAll is also updated to use the new library for native GPIO pins.
func GPIOOutAll(name string, command string) {
	for _, io := range Config.Global.Hardware.IO.Pins.Pin {
		if io.Enabled && io.Direction == "output" && io.Device == "led/relay" {
			switch io.Type {
			case "gpio":
				line, ok := outputLines[io.Name]
				if !ok {
					continue // Skip if not initialized
				}
				val := 0
				if command == "on" {
					val = 1
					if io.Inverted {
						val = 0
					}
				} else { // "off"
					if io.Inverted {
						val = 1
					}
				}
				_ = line.SetValue(val)
			default:
				log.Println("error: GPIO Types Currently Supported are gpio or mcp23017 only!")
			}
		}
	}
}

// --- The remaining functions are unaffected by the GPIO library change and are included as-is. ---

func Max7219(max7219Cascaded int, spiBus int, spiDevice int, brightness byte, toDisplay string) {
	if Config.Global.Hardware.IO.Max7219.Enabled {
		mtx := max7219.NewMatrix(max7219Cascaded)
		err := mtx.Open(spiBus, spiDevice, brightness)
		if err != nil {
			log.Fatal(err)

		}
		mtx.Device.SevenSegmentDisplay(toDisplay)
		defer mtx.Close()
	}
}

func (b *Talkkonnect) rotaryAction(direction string) {
	if Config.Global.Hardware.IO.RotaryEncoder.Enabled {
		if direction == "cw" {
			log.Println("debug: Rotating Clockwise")
			switch RotaryFunction.Function {
			case "mumblechannel":
				if b.findEnabledRotaryEncoderFunction("mumblechannel") {
					b.ChannelUp()
				}
			case "localvolume":
				if b.findEnabledRotaryEncoderFunction("localvolume") {
					b.cmdVolumeRXUp()
				}
			case "radiochannel":
				if b.findEnabledRotaryEncoderFunction("radiochannel") {
					go radioChannelIncrement("up")
				}
			case "voicetarget":
				if b.findEnabledRotaryEncoderFunction("voicetarget") {
					b.VTMove("up")
				}
			default:
				log.Println("error: No Rotary Function Enabled in Config")
				return
			}
			playIOMedia("iorotarycw")
		}
		if direction == "ccw" {
			log.Println("debug: Rotating CounterClockwise")
			switch RotaryFunction.Function {
			case "mumblechannel":
				if b.findEnabledRotaryEncoderFunction("mumblechannel") {
					b.ChannelDown()
				}
			case "localvolume":
				if b.findEnabledRotaryEncoderFunction("localvolume") {
					b.cmdVolumeRXDown()
				}
			case "radiochannel":
				if b.findEnabledRotaryEncoderFunction("radiochannel") {
					go radioChannelIncrement("down")
				}
			case "voicetarget":
				if b.findEnabledRotaryEncoderFunction("voicetarget") {
					b.VTMove("down")
				}
			default:
				log.Println("error: No Rotary Function Enabled in Config")
				return
			}
			playIOMedia("iorotaryccw")
		}
	}
}

func createEnabledRotaryEncoderFunctions() {
	for item, control := range Config.Global.Hardware.IO.RotaryEncoder.Control {
		if control.Enabled {
			RotaryFunctions = append(RotaryFunctions, rotaryFunctionsStruct{item, control.Function})
		}
	}
}

// here
func (b *Talkkonnect) nextEnabledRotaryEncoderFunction() {
	if len(RotaryFunctions) > RotaryFunction.Item+1 {
		RotaryFunction.Item++
		RotaryFunction.Function = RotaryFunctions[RotaryFunction.Item].Function
		log.Printf("info: Current Rotary Item %v Function %v\n", RotaryFunction.Item, RotaryFunction.Function)
		if RotaryFunction.Function == "mumblechannel" {
			b.sevenSegment("mumblechannel", strconv.Itoa(int(b.Client.Self.Channel.ID)))
			if OLEDEnabled {
				oledDisplay(false, 6, OLEDStartColumn, "[Channel Mode]      ")
			}
		}
		if RotaryFunction.Function == "localvolume" {
			b.cmdCurrentRXVolume()
			if OLEDEnabled {
				oledDisplay(false, 6, OLEDStartColumn, "[Volume Mode]       ")
			}
		}
		if RotaryFunction.Function == "radiochannel" {
			b.sevenSegment("radiochannel", "")
			if OLEDEnabled {
				oledDisplay(false, 6, OLEDStartColumn, "[Radio Channel Mode]")
			}
		}
		if RotaryFunction.Function == "voicetarget" {
			b.sevenSegment("voicetarget", "")
			if OLEDEnabled {
				oledDisplay(false, 6, OLEDStartColumn, "[Voice Target Mode] ")
			}
		}
		return
	}

	if len(RotaryFunctions) == RotaryFunction.Item+1 {
		RotaryFunction.Item = 0
		RotaryFunction.Function = RotaryFunctions[0].Function
		log.Printf("info: Current Rotary Item %v Function %v\n", RotaryFunction.Item, RotaryFunction.Function)
		if RotaryFunction.Function == "mumblechannel" {
			b.sevenSegment("mumblechannel", strconv.Itoa(int(b.Client.Self.Channel.ID)))
			if OLEDEnabled {
				oledDisplay(false, 6, OLEDStartColumn, "[Channel Mode]      ")
			}
		}
		if RotaryFunction.Function == "localvolume" {
			b.cmdCurrentRXVolume()
			if OLEDEnabled {
				oledDisplay(false, 6, OLEDStartColumn, "[Volume Mode]       ")
			}
		}
		if RotaryFunction.Function == "radiochannel" {
			b.sevenSegment("radiochannel", "")
			if OLEDEnabled {
				oledDisplay(false, 6, OLEDStartColumn, "[Radio Channel Mode]")
			}
		}
		if RotaryFunction.Function == "voicetarget" {
			b.sevenSegment("voicetarget", "")
			if OLEDEnabled {
				oledDisplay(false, 6, OLEDStartColumn, "[Voice Target Mode] ")
			}
		}
		return
	}
}

func (b *Talkkonnect) findEnabledRotaryEncoderFunction(findFunction string) bool {
	for _, functionName := range Config.Global.Hardware.IO.RotaryEncoder.Control {
		if findFunction == functionName.Function {
			return functionName.Enabled
		}
	}
	return false
}

func GPIOOutputPinControl(name string, command string) {
	if Config.Global.Hardware.TargetBoard != "rpi" {
		return
	}
	for i, io := range Config.Global.Hardware.IO.Pins.Pin {
		if io.Direction == "output" && io.Name == name {
			switch command {
			case "off":
				Config.Global.Hardware.IO.Pins.Pin[i].Enabled = false
			case "on":
				Config.Global.Hardware.IO.Pins.Pin[i].Enabled = true
			case "toggle":
				Config.Global.Hardware.IO.Pins.Pin[i].Enabled = !Config.Global.Hardware.IO.Pins.Pin[i].Enabled
			}
			log.Printf("GPIO Enabled For Pin %v is Now Set To %v\n", io.Name, Config.Global.Hardware.IO.Pins.Pin[i].Enabled)
		}
	}
}

func GPIOInputPinControl(name string, command string) {
	if Config.Global.Hardware.TargetBoard != "rpi" {
		return
	}

	for _, io := range Config.Global.Hardware.IO.Pins.Pin {
		if io.Direction == "input" {
			if io.Name == "txptt" && io.Name == name {
				switch command {
				case "off":
					TxButtonUsed = false
				case "on":
					TxButtonUsed = true
				case "toggle":
					TxButtonUsed = !TxButtonUsed
				}
				log.Printf("%v Enabled is Now Set To %v\n", io.Name, TxButtonUsed)
			}
			if io.Name == "txtoggle" && io.Name == name {
				switch command {
				case "off":
					TxToggleUsed = false
				case "on":
					TxToggleUsed = true
				case "toggle":
					TxToggleUsed = !TxToggleUsed
				}
				log.Printf("%v Enabled is Now Set To %v\n", io.Name, TxToggleUsed)
			}
			if io.Name == "channelup" && io.Name == name {
				switch command {
				case "off":
					UpButtonUsed = false
				case "on":
					UpButtonUsed = true
				case "toggle":
					UpButtonUsed = !UpButtonUsed
				}
				log.Printf("%v Enabled is Now Set To %v\n", io.Name, UpButtonUsed)
			}
			if io.Name == "channeldown" && io.Name == name {
				switch command {
				case "off":
					DownButtonUsed = false
				case "on":
					DownButtonUsed = true
				case "toggle":
					DownButtonUsed = !DownButtonUsed
				}
				log.Printf("%v Enabled is Now Set To %v\n", io.Name, DownButtonUsed)
			}
			if io.Name == "panic" && io.Name == name {
				switch command {
				case "off":
					PanicUsed = false
				case "on":
					PanicUsed = true
				case "toggle":
					PanicUsed = !PanicUsed
				}
				log.Printf("%v Enabled is Now Set To %v\n", io.Name, PanicUsed)
			}
			if io.Name == "streamtoggle" && io.Name == name {
				switch command {
				case "off":
					StreamToggleUsed = false
				case "on":
					StreamToggleUsed = true
				case "toggle":
					StreamToggleUsed = !StreamToggleUsed
				}
				log.Printf("%v Enabled is Now Set To %v\n", io.Name, StreamToggleUsed)
			}
			if io.Name == "comment" && io.Name == name {
				switch command {
				case "off":
					CommentUsed = false
				case "on":
					CommentUsed = true
				case "toggle":
					CommentUsed = !CommentUsed
				}
				log.Printf("%v Enabled is Now Set To %v\n", io.Name, CommentUsed)
			}
			if io.Name == "listening" && io.Name == name {
				switch command {
				case "off":
					ListeningUsed = false
				case "on":
					ListeningUsed = true
				case "toggle":
					ListeningUsed = !ListeningUsed
				}
				log.Printf("%v Enabled is Now Set To %v\n", io.Name, ListeningUsed)
			}
			if io.Name == "rotarybutton" && io.Name == name {
				switch command {
				case "off":
					RotaryButtonUsed = false
				case "on":
					RotaryButtonUsed = true
				case "toggle":
					RotaryButtonUsed = !RotaryButtonUsed
				}
				log.Printf("%v Enabled is Now Set To %v\n", io.Name, RotaryButtonUsed)
			}
			if io.Name == "volup" && io.Name == name {
				switch command {
				case "off":
					VolUpButtonUsed = false
				case "on":
					VolUpButtonUsed = true
				case "toggle":
					VolUpButtonUsed = !VolUpButtonUsed
				}
				log.Printf("%v Enabled is Now Set To %v\n", io.Name, VolUpButtonUsed)
			}
			if io.Name == "voldown" && io.Name == name {
				switch command {
				case "off":
					VolDownButtonUsed = false
				case "on":
					VolDownButtonUsed = true
				case "toggle":
					VolDownButtonUsed = !VolDownButtonUsed
				}
				log.Printf("%v Enabled is Now Set To %v\n", io.Name, VolDownButtonUsed)
			}
			if io.Name == "tracking" && io.Name == name {
				switch command {
				case "off":
					TrackingUsed = false
				case "on":
					TrackingUsed = true
				case "toggle":
					TrackingUsed = !TrackingUsed
				}
				log.Printf("%v Enabled is Now Set To %v\n", io.Name, TrackingUsed)
			}
			if io.Name == "mqtt0" && io.Name == name {
				switch command {
				case "off":
					MQTT0ButtonUsed = false
				case "on":
					MQTT0ButtonUsed = true
				case "toggle":
					MQTT0ButtonUsed = !MQTT0ButtonUsed
				}
				log.Printf("%v Enabled is Now Set To %v\n", io.Name, MQTT0ButtonUsed)
			}
			if io.Name == "mqtt1" && io.Name == name {
				switch command {
				case "off":
					MQTT1ButtonUsed = false
				case "on":
					MQTT1ButtonUsed = true
				case "toggle":
					MQTT1ButtonUsed = !MQTT0ButtonUsed
				}
				log.Printf("%v Enabled is Now Set To %v\n", io.Name, MQTT1ButtonUsed)
			}
			if io.Name == "nextserver" && io.Name == name {
				switch command {
				case "off":
					NextServerButtonUsed = false
				case "on":
					NextServerButtonUsed = true
				case "toggle":
					NextServerButtonUsed = !NextServerButtonUsed
				}
				log.Printf("%v Enabled is Now Set To %v\n", io.Name, NextServerButtonUsed)
			}
			if io.Name == "repeatertone" && io.Name == name {
				switch command {
				case "off":
					RepeaterToneButtonUsed = false
				case "on":
					RepeaterToneButtonUsed = true
				case "toggle":
					RepeaterToneButtonUsed = !RepeaterToneButtonUsed
				}
				log.Printf("%v Enabled is Now Set To %v\n", io.Name, RepeaterToneButtonUsed)
			}
			if io.Name == "memorychannel1" && io.Name == name {
				switch command {
				case "off":
					MemoryChannelButton1Used = false
				case "on":
					MemoryChannelButton1Used = true
				case "toggle":
					MemoryChannelButton1Used = !MemoryChannelButton1Used
				}
				log.Printf("%v Enabled is Now Set To %v\n", io.Name, MemoryChannelButton1Used)
			}
			if io.Name == "memorychannel2" && io.Name == name {
				switch command {
				case "off":
					MemoryChannelButton2Used = false
				case "on":
					MemoryChannelButton2Used = true
				case "toggle":
					MemoryChannelButton2Used = !MemoryChannelButton2Used
				}
				log.Printf("%v Enabled is Now Set To %v\n", io.Name, MemoryChannelButton2Used)
			}
			if io.Name == "memorychannel3" && io.Name == name {
				switch command {
				case "off":
					MemoryChannelButton3Used = false
				case "on":
					MemoryChannelButton3Used = true
				case "toggle":
					MemoryChannelButton3Used = !MemoryChannelButton3Used
				}
				log.Printf("%v Enabled is Now Set To %v\n", io.Name, MemoryChannelButton3Used)
			}
			if io.Name == "memorychannel4" && io.Name == name {
				switch command {
				case "off":
					MemoryChannelButton4Used = false
				case "on":
					MemoryChannelButton4Used = true
				case "toggle":
					MemoryChannelButton4Used = !MemoryChannelButton4Used
				}
			}
			if io.Name == "presetvoicetarget2" && io.Name == name {
				switch command {
				case "off":
					VoiceTargetButton2Used = false
				case "on":
					VoiceTargetButton2Used = true
				case "toggle":
					VoiceTargetButton2Used = !VoiceTargetButton2Used
				}
			}
			if io.Name == "presetvoicetarget3" && io.Name == name {
				switch command {
				case "off":
					VoiceTargetButton3Used = false
				case "on":
					VoiceTargetButton3Used = true
				case "toggle":
					VoiceTargetButton3Used = !VoiceTargetButton3Used
				}
			}
			if io.Name == "presetvoicetarget4" && io.Name == name {
				switch command {
				case "off":
					VoiceTargetButton4Used = false
				case "on":
					VoiceTargetButton4Used = true
				case "toggle":
					VoiceTargetButton4Used = !VoiceTargetButton4Used
				}
			}
			if io.Name == "presetvoicetarget5" && io.Name == name {
				switch command {
				case "off":
					VoiceTargetButton5Used = false
				case "on":
					VoiceTargetButton5Used = true
				case "toggle":
					VoiceTargetButton5Used = !VoiceTargetButton5Used
				}
			}
		}
	}
}

func analogZone(announcementChannel string, IOName string) {
	go func() {
		var lastChannel string = ""
		for {
			select {
			case f := <-Talking:
				if (f.OnChannel == announcementChannel) && (lastChannel != announcementChannel) {
					go GPIOOutPin(IOName, "on")
					lastChannel = f.OnChannel
				}
			case <-TalkedTicker.C:
				if lastChannel == announcementChannel {
					go GPIOOutPin(IOName, "off")
					lastChannel = ""
				}
			}
		}
	}()
}

func analogCreateZones() {
	if Config.Global.Hardware.TargetBoard != "rpi" {
		return
	}

	if !Config.Global.Hardware.AnalogRelays.Enabled {
		log.Printf("debug: Skipping the Creation of Analog Zones\n")
		return
	}

	for i, io := range Config.Global.Hardware.AnalogRelays.Zones.Zone {
		if io.Enabled {
			for _, ii := range Config.Global.Hardware.AnalogRelays.Zones.Zone[i].Pins.Name {
				analogZone(io.ListenChannel, ii)
				log.Printf("debug: Creating Analog Zones For Zone %v Relays %v\n", io.Name, ii)
			}
		}
	}
}
