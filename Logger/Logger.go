package Logger

import (
	"fmt"
	"os"

	// "runtime"
	"strings"
	"time"
)

var txt_file_for_logger *os.File
var txt_file_count = 0
var time_last_logger_file int64
var log_enable_flag string

func Logger(map_message string, err interface{}, flag ...string) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("ERROR : LOGGER ", err)
		}
	}()
	if len(flag) > 0 {
		log_enable_flag = flag[0]
	}
	if log_enable_flag == "true" {
		time_now := time.Now().Unix()
		if txt_file_count == 0 || time_now-time_last_logger_file > 60*60 {
			time_last_logger_file = time.Now().Unix()
			now_time := strings.Split(time.Now().String(), " ")
			now_time_data := strings.Replace(strings.Split(now_time[0]+" "+now_time[1], ".")[0], ":", "-", 2)
			txt_file_for_logger, _ = os.Create("Logs/" + now_time_data + ".txt")
			txt_file_count += 1
		}
		txt_file_for_logger.WriteString(map_message + "\n")
		// fmt.Println(runtime.NumGoroutine())
		if err != nil {
			func() {
				defer func() {
					txt_file_for_logger.WriteString(err.(string) + "\n")
				}()
				txt_file_for_logger.WriteString(err.(error).Error() + "\n")
			}()

		}
	}
	fmt.Println(map_message)
	if err != nil {
		fmt.Println(err)
	}
}
