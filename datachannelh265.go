package webrtc

import (
	"bytes"
	"encoding/binary"
	"errors"
	"strconv"
	// "fmt"
	. "github.com/pion/webrtc/v3"
)

const (
	//H265
	// https://zhuanlan.zhihu.com/p/458497037
	NALU_H265_VPS       = 0x4001
	NALU_H265_SPS       = 0x4201
	NALU_H265_PPS       = 0x4401
	NALU_H265_SEI       = 0x4e01
	NALU_H265_IFRAME    = 0x2601
	NALU_H265_PFRAME    = 0x0201
	HEVC_NAL_TRAIL_N    = 0
	HEVC_NAL_TRAIL_R    = 1
	HEVC_NAL_TSA_N      = 2
	HEVC_NAL_TSA_R      = 3
	HEVC_NAL_STSA_N     = 4
	HEVC_NAL_STSA_R     = 5
	HEVC_NAL_BLA_W_LP   = 16
	HEVC_NAL_BLA_W_RADL = 17
	HEVC_NAL_BLA_N_LP   = 18
	HEVC_NAL_IDR_W_RADL = 19
	HEVC_NAL_IDR_N_LP   = 20
	HEVC_NAL_CRA_NUT    = 21
	HEVC_NAL_RADL_N     = 6
	HEVC_NAL_RADL_R     = 7
	HEVC_NAL_RASL_N     = 8
	HEVC_NAL_RASL_R     = 9
	MAXPACKETSIZE       = 65536
)


func SendH265FrameData(dc *DataChannel, data []byte,timestamp int64) {
	if len(data) > 4 && dc != nil && dc.ReadyState() == DataChannelStateOpen {
		var frametypestr string
		glength := len(data)
		count := glength / MAXPACKETSIZE
		rem := glength % MAXPACKETSIZE
		packets := count
		if(rem != 0){
			packets++
		}
		temptype, frametype, err := GetFrameType(data)
		if err != nil {
 
		} else {
			frametypestr, err = GetFrameTypeName(frametype)
		}

		startstr := "h265 start ,FrameType:" + frametypestr + ",nalutype:" + strconv.Itoa(int(temptype)) + ",pts:" + strconv.FormatInt(timestamp, 10) + ",Packetslen:" + strconv.Itoa(glength) + ",packets:" + strconv.Itoa(packets) + ",rem:" + strconv.Itoa(rem)

		dc.SendText(startstr)
		i := 0
		for i = 0; i < count; i++ {
			lenth := i * MAXPACKETSIZE
			dc.Send(data[lenth : lenth+MAXPACKETSIZE])
		}
		if rem != 0 {
			dc.Send(data[glength-rem : glength])
		}
		dc.SendText("h265 end")

	}


}

func GetFrameType(pdata []byte) (uint8, uint16, error) {
	var frametype uint16
 
	destcount := 0
	if FindStartCode2(pdata) {
		destcount = 3
	} else if FindStartCode3(pdata) {
		destcount = 4
	} else {
		return 0, 0, errors.New("not find")
		destcount = 4
	}
	temptype := (pdata[destcount] & 0x7E) >> 1
	bytesBuffer := bytes.NewBuffer(pdata[destcount : destcount+2])
	binary.Read(bytesBuffer, binary.BigEndian, &frametype)
	return temptype, frametype, nil
}

func GetFrameTypeName(frametype uint16) (string, error) {
	switch frametype {
	case NALU_H265_VPS:
		return "H265_FRAME_VPS", nil
	case NALU_H265_SPS:
		return "H265_FRAME_SPS", nil
	case NALU_H265_PPS:
		return "H265_FRAME_PPS", nil
	case NALU_H265_SEI:
		return "H265_FRAME_SEI", nil
	case NALU_H265_IFRAME:
		return "H265_FRAME_I", nil
	case NALU_H265_PFRAME:
		return "H265_FRAME_P", nil
	default:
		return "", errors.New("frametype unsupport")
	}
}

func FindStartCode2(Buf []byte) bool {
	if Buf[0] != 0 || Buf[1] != 0 || Buf[2] != 1 {
		return false //判断是否为0x000001,如果是返回1
	} else {
		return true
	}
}
 
func FindStartCode3(Buf []byte) bool {
	if Buf[0] != 0 || Buf[1] != 0 || Buf[2] != 0 || Buf[3] != 1 {
		return false //判断是否为0x00000001,如果是返回1
	} else {
		return true
	}
}

func Add3ZoneOne(h265frame []byte) []byte{
	var hBuf = [4]byte{0, 0, 0, 1}
	var data []byte
    for i := range hBuf {
        data = append(data, byte(hBuf[i]))
    }
    for i := range h265frame {
        data = append(data, byte(h265frame[i]))
    }
    return data
}

func AddBufs(A []byte,B []byte) []byte{
	var data []byte
    for i := range A {
        data = append(data, byte(A[i]))
    }
    for i := range B {
        data = append(data, byte(B[i]))
    }
    return data
}



type WebRtcReturn struct {
	SessionDescription
	IsH265 bool `json:"isH265"`
}