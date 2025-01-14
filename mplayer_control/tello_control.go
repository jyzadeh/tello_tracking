package main

import (
	"fmt"
	"os/exec"
	"time"

	"gobot.io/x/gobot"
	"gobot.io/x/gobot/platforms/dji/tello"
	"gobot.io/x/gobot/platforms/joystick"
)

// Maximum joystick axis value
const maxJoyVal = 32768

// Drone driver
var drone = tello.NewDriver("8889")

// Joystick adapter
var joyAdaptor = joystick.NewAdaptor()

// Joystick driver
var stick = joystick.NewDriver(joyAdaptor, "dualshock4")

// Drone flight data
var flightData *tello.FlightData

// Drone wifi data
var wifiData *tello.WifiData

func main() {
	establishConnection()

	prepareVideo()

	handleJoystick()

	robot := gobot.NewRobot("tello",
		[]gobot.Connection{},
		[]gobot.Connection{joyAdaptor},
		[]gobot.Device{drone},
		[]gobot.Device{stick},
	)

	robot.Start()
}

// Handles joystick input
func handleJoystick() {
	stick.On(joystick.OptionsPress, func(data interface{}) {
		speed, _ := drone.ChangeSpeed()
		if speed == 0 {
			fmt.Println("Set drone to low speed")
		} else {
			fmt.Println("Set drone to high speed")
		}
	})
	stick.On(joystick.TouchpadPress, func(data interface{}) {
		drone.Bounce()
	})
	stick.On(joystick.SquarePress, func(data interface{}) {
		drone.ThrowTakeOff()
	})
	stick.On(joystick.CirclePress, func(data interface{}) {
		fmt.Println("Battery:", flightData.BatteryPercentage)
		fmt.Println("Wifi:", wifiData.Strength, wifiData.Disturb)
	})
	stick.On(joystick.TrianglePress, func(data interface{}) {
		drone.TakeOff()
		println("Takeoff")
	})
	stick.On(joystick.XPress, func(data interface{}) {
		drone.Land()
		println("Land")
	})
	stick.On(joystick.RightY, func(data interface{}) {
		val := float64(data.(int16))
		if val >= 0 {
			drone.Backward(tello.ValidatePitch(val, maxJoyVal))
		} else {
			drone.Forward(tello.ValidatePitch(val, maxJoyVal))
		}
	})
	stick.On(joystick.RightX, func(data interface{}) {
		val := float64(data.(int16))
		if val >= 0 {
			drone.Right(tello.ValidatePitch(val, maxJoyVal))
		} else {
			drone.Left(tello.ValidatePitch(val, maxJoyVal))
		}
	})
	stick.On(joystick.LeftY, func(data interface{}) {
		val := float64(data.(int16))
		if val >= 0 {
			drone.Down(tello.ValidatePitch(val, maxJoyVal))
		} else {
			drone.Up(tello.ValidatePitch(val, maxJoyVal))
		}
	})
	stick.On(joystick.LeftX, func(data interface{}) {
		val := float64(data.(int16))
		if val >= 0 {
			drone.Clockwise(tello.ValidatePitch(val, maxJoyVal))
		} else {
			drone.CounterClockwise(tello.ValidatePitch(val, maxJoyVal))
		}
	})
	stick.On(joystick.UpPress, func(data interface{}) {
		drone.FrontFlip()
	})
	stick.On(joystick.DownPress, func(data interface{}) {
		drone.BackFlip()
	})
	stick.On(joystick.RightPress, func(data interface{}) {
		drone.RightFlip()
	})
	stick.On(joystick.LeftPress, func(data interface{}) {
		drone.LeftFlip()
	})
}

// Prepares the drone video feed to output through mplayer
func prepareVideo() {
	mplayer := exec.Command("mplayer", "-fps", "60", "-xy", "500", "-")
	mplayerIn, _ := mplayer.StdinPipe()
	if err := mplayer.Start(); err != nil {
		fmt.Println(err)
		return
	}
	drone.On(tello.VideoFrameEvent, func(data interface{}) {
		pkt := data.([]byte)
		if _, err := mplayerIn.Write(pkt); err != nil {
			println(err)
		}
	})
}

// Establishes the connection to the drone and initializes continual fetching of flight data
func establishConnection() {
	drone.On(tello.FlightDataEvent, func(data interface{}) {
		flightData = data.(*tello.FlightData)
	})
	drone.On(tello.WifiDataEvent, func(data interface{}) {
		wifiData = data.(*tello.WifiData)
	})
	drone.On(tello.ConnectedEvent, func(data interface{}) {
		fmt.Println("Connected")
		drone.StartVideo()
		drone.SetVideoEncoderRate(4)
		gobot.Every(100*time.Millisecond, func() {
			drone.StartVideo()
		})
	})
}
