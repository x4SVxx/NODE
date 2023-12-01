package ReportsAndMessages

import (
	"encoding/binary"
	"encoding/hex"
	"math"
)

func GetIntFrom8Bytes(mas_bytes []byte) int {
	var new_mas_bytes [8]byte
	copy(new_mas_bytes[0:len(mas_bytes)], mas_bytes)
	return int(binary.LittleEndian.Uint64(new_mas_bytes[:]))
}

func DecodeAnchorMessage(data []byte) map[string]interface{} {
	FnCE := data[0]
	msg := map[string]interface{}{}
	if FnCE == 48 {
		msg["type"] = "CS_TX"
		msg["sender"] = " "
		msg["receiver"] = " "
		msg["seq"] = int(data[1])
		msg["timestamp"] = float64(GetIntFrom8Bytes(data[2:7])) * (1.0 / 499.2e6 / 128.0)
	} else if FnCE == 49 {
		msg["type"] = "CS_RX"
		msg["sender"] = string(hex.EncodeToString([]byte{data[3]})) + string(hex.EncodeToString([]byte{data[2]}))
		msg["receiver"] = " "
		msg["seq"] = int(data[1])
		msg["timestamp"] = float64(GetIntFrom8Bytes(data[10:15])) * (1.0 / 499.2e6 / 128.0)
	} else if FnCE == 50 {
		msg["type"] = "BLINK"
		msg["sender"] = string(hex.EncodeToString([]byte{data[3]})) + string(hex.EncodeToString([]byte{data[2]}))
		msg["receiver"] = " "
		msg["sn"] = int(data[1])
		msg["timestamp"] = float64(GetIntFrom8Bytes(data[10:15])) * (1.0 / 499.2e6 / 128.0)
	} else if FnCE == 52 {
		msg["type"] = "BLINK"
		msg["sender"] = string(hex.EncodeToString([]byte{data[3]})) + string(hex.EncodeToString([]byte{data[2]}))
		msg["receiver"] = " "
		msg["sn"] = int(data[1])
		msg["state"] = data[20]
		msg["timestamp"] = float64(GetIntFrom8Bytes(data[10:15])) * (1.0 / 499.2e6 / 128.0)
		msg["tx_timestamp"] = float64(GetIntFrom8Bytes(data[15:20])) * (1.0 / 499.2e6 / 128.0)
	} else if FnCE == 53 {
		msg["type"] = "BLINK"
		msg["sender"] = string(hex.EncodeToString([]byte{data[3]})) + string(hex.EncodeToString([]byte{data[2]}))
		msg["receiver"] = " "
		msg["sn"] = int(data[1])
		msg["state"] = data[15]
		msg["timestamp"] = float64(GetIntFrom8Bytes(data[10:15])) * (1.0 / 499.2e6 / 128.0)
		msg["tx_timestamp"] = " "
	} else if FnCE == 66 {
		msg["type"] = "Config request"
		msg["receiver"] = string(hex.EncodeToString([]byte{data[2]})) + string(hex.EncodeToString([]byte{data[1]}))
	} else {
		msg["type"] = "Unknow"
	}
	return msg
}

func Build_RTLS_START_REQ(ON_OFF int) []byte {
	return []byte{
		byte(0x57),   // 1 byte
		byte(ON_OFF), // 1 byte
	}
}

func Build_RTLS_CMD_SET_CFG_CCP(M int, CP int, PRF int, DR int, PC int, PL int, PSN_L int, PSN_U int, ADRx int, ADTx int, LD int, Lag int) []byte {
	ADRx_bytes_mas := make([]byte, 2)
	binary.LittleEndian.PutUint16(ADRx_bytes_mas, uint16(ADRx))
	ADTx_bytes_mas := make([]byte, 2)
	binary.LittleEndian.PutUint16(ADTx_bytes_mas, uint16(ADTx))
	null_bytes_mas := make([]byte, 8)
	binary.LittleEndian.PutUint16(null_bytes_mas, uint16(0))
	Lag_bytes_mas := make([]byte, 4)
	binary.LittleEndian.PutUint16(Lag_bytes_mas, uint16(Lag))
	return []byte{
		byte(0x44), // 1 byte
		byte(0),    // 1 byte
		byte(M),    // 1 byte
		byte(int(float64(CP) + float64(PRF)*math.Pow(2, 4))), // 1 byte
		byte(DR), // 1 byte
		byte(PC), // 1 byte
		byte(PL), // 1 byte
		byte(int(float64(PSN_L) + float64(PSN_U)*math.Pow(2, 4))), // 1 byte
		ADRx_bytes_mas[0], // 2 bytes
		ADRx_bytes_mas[1],
		ADTx_bytes_mas[0], // 2 bytes
		ADTx_bytes_mas[1],
		byte(0),           // 1 byte
		byte(LD),          // 1 byte
		null_bytes_mas[0], // 8 bytes
		null_bytes_mas[1],
		null_bytes_mas[2],
		null_bytes_mas[3],
		null_bytes_mas[4],
		null_bytes_mas[5],
		null_bytes_mas[6],
		null_bytes_mas[7],
		Lag_bytes_mas[0], // 4 bytes
		Lag_bytes_mas[1],
		Lag_bytes_mas[2],
		Lag_bytes_mas[3],
	}
}
