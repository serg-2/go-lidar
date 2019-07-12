package main

import "fmt"
import "encoding/hex"
import "log"
import "time"

//Make Exit on Ctrl +C
import "os"
import "os/signal"
import "syscall"

const (
        I2C_SLAVE = 0x0703
)

// I2Cstruc represents a connection to I2C-device.
type I2Cstruc struct {
        addr uint8
        bus  int
        rc   *os.File
}

func ioctl(fd, cmd, arg uintptr) error {
        _, _, err := syscall.Syscall6(syscall.SYS_IOCTL, fd, cmd, arg, 0, 0, 0)
        if err != 0 {
                return err
        }
        return nil
}

// NewI2C opens a connection for I2C-device.
// SMBus (System Management Bus) protocol over I2C
// supported as well: you should preliminary specify
// register address to read from, either write register
// together with the data in case of write operations.
func NewI2Cdevice(addr uint8, bus int) (*I2Cstruc, error) {
        f, err := os.OpenFile(fmt.Sprintf("/dev/i2c-%d", bus), os.O_RDWR, 0600)
        if err != nil {
                return nil, err
        }
        if err := ioctl(f.Fd(), I2C_SLAVE, uintptr(addr)); err != nil {
                return nil, err
        }
        v := &I2Cstruc{rc: f, bus: bus, addr: addr}
        return v, nil
}

// Write sends bytes to the remote I2C-device. The interpretation of
// the message is implementation-dependant.
func (v *I2Cstruc) WriteBytes(buf []byte) (int, error) {
        return v.write(buf)
}

func (v *I2Cstruc) write(buf []byte) (int, error) {
        return v.rc.Write(buf)
}

func (v *I2Cstruc) read(buf []byte) (int, error) {
	return v.rc.Read(buf)
}

// ReadBytes reads bytes from I2C-device.
// Number of bytes read correspond to buf parameter length.
func (v *I2Cstruc) ReadBytes(buf []byte) (int, error) {
	n, err := v.read(buf)
	if err != nil {
		return n, err
	}
	return n, nil
}

// Close I2C-connection.
func (v *I2Cstruc) Close() error {
        return v.rc.Close()
}

func main () {

fmt.Println("Starting")
a,_ := NewI2Cdevice(0x10,1)

// Query - 'D' Hex code 0x44
query_code,err := hex.DecodeString("44")
if err != nil {
  log.Fatal(err)
}

// Reboot - 'X' Hex code 0x58
reboot,err := hex.DecodeString("58")
if err != nil {
  log.Fatal(err)
}

// Turn Off LED - 'E' Hex code 0x45
//led_turnoff,err := hex.DecodeString("45")
//if err != nil {
//  log.Fatal(err)
//}

// Mode Continuous - 'MC' Hex code 0x4d 0x43
//mode_cont,err := hex.DecodeString("4d43")
//if err != nil {
//  log.Fatal(err)
//}

tick_sec := time.Tick(1* time.Second)
//tick_min := time.Tick(60* time.Second)

//Activating Ctrl+C
quit := make(chan os.Signal)
signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
cleanupDone := make(chan struct{})
go func() {
    <-quit
    close(cleanupDone)
}()
buf := make([]byte, 20)

_,err=a.WriteBytes(reboot)
      if err != nil {
         fmt.Println("Check that address i2c 0x10 is not in use. (By pigpio daemon for example)")
         log.Fatal(err)
      }

time.Sleep(1000 * time.Millisecond)

// _,err=a.WriteBytes(led_turnoff)
//      if err != nil {
//         log.Fatal(err)
//      }

//fmt.Println("Changing mode to continuos")
//_,err=a.WriteBytes(mode_cont)
//     if err != nil {
//         log.Fatal(err)
//      }


//Main CYCLE
for {
  select {
    case <-cleanupDone:
      a.Close()
      fmt.Println("Finished")
      os.Exit(1)
    case <-tick_sec:
      _,err=a.WriteBytes(query_code)
      if err != nil {
         fmt.Println("Error during query")
         log.Fatal(err)
      }
      _,err=a.ReadBytes(buf)
      if err != nil {
        fmt.Println("Error during Read")
        log.Fatal(err)
      }
      fmt.Printf("\rRange: %d mm    ", int(buf[0])*256 + int(buf[1]) )
    default:
      // Lidar frequency up to 60Hz  
      // time.Sleep(17 * time.Millisecond)
      // Lidar High accuracy mode up to 5Hz
      time.Sleep(200 * time.Millisecond)
  }

}
}
