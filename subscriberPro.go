package webrtc

import (
	"fmt"
	"strings"

	"github.com/pion/rtcp"
	. "github.com/pion/webrtc/v3"
	. "m7s.live/engine/v4"
	"m7s.live/engine/v4/common"
	"time"
	"m7s.live/engine/v4/codec"
	"m7s.live/engine/v4/track"
)

type WebRTCSubscriberPro struct {
	Subscriber
	WebRTCIO
	videoTrack *TrackLocalStaticRTP
	audioTrack *TrackLocalStaticRTP
	isH265 bool
}

func (suber *WebRTCSubscriberPro) OnEvent(event any) {
	switch v := event.(type) {
	case *track.Video:
		if v.CodecID == codec.CodecID_H264 {
			suber.isH265 = false
			pli := "42001f"
			pli = fmt.Sprintf("%x", v.GetDecoderConfiguration().Raw[0][1:4])
			if !strings.Contains(suber.SDP, pli) {
				pli = reg_level.FindAllStringSubmatch(suber.SDP, -1)[0][1]
			}
			suber.videoTrack, _ = NewTrackLocalStaticRTP(RTPCodecCapability{MimeType: MimeTypeH264, SDPFmtpLine: "level-asymmetry-allowed=1;packetization-mode=1;profile-level-id=" + pli}, "video", "m7s")
			rtpSender, _ := suber.PeerConnection.AddTrack(suber.videoTrack)
			go func() {
				rtcpBuf := make([]byte, 1500)
				for {
					if n, _, rtcpErr := rtpSender.Read(rtcpBuf); rtcpErr != nil {

						return
					} else {
						if p, err := rtcp.Unmarshal(rtcpBuf[:n]); err == nil {
							for _, pp := range p {
								switch pp.(type) {
								case *rtcp.PictureLossIndication:
									// fmt.Println("PictureLossIndication")
								}
							}
						}
					}
				}
			}()
			suber.Subscriber.AddTrack(v) //接受这个track
		}

		start := time.Now().UnixMilli()
		var rtcDc *DataChannel
		if v.CodecID == codec.CodecID_H265 {
			suber.isH265 = true
			nInSendH265Track :=0
			suber.PeerConnection.OnDataChannel(func(dc *DataChannel) {
				rtcDc = dc
				rtcDc.OnOpen(func() {
					va := v.IDRing.Value
					fmt.Printf("dc.OnOpen %d\n", nInSendH265Track)

					var h265frame []byte

					if va.IFrame { //如果是关键帧,增加关键帧头
						h265frame = []byte{0,0,0,1,64,1,12,1,255,255,1,96,0,0,3,0,176,0,0,3,0,0,3,0,123,172,12,0,0,56,64,0,5,126,66,168,0,0,0,1,66,1,1,1,96,0,0,3,0,176,0,0,3,0,0,3,0,123,160,3,192,128,16,229,141,174,228,203,243,112,16,16,16,64,0,3,132,0,0,87,228,40,64,0,0,0,1,68,1,192,242,176,59,36}
					}

					for _, packets := range va.Raw {
						for _, packet := range packets {
							if len(h265frame)==0 {
								h265frame = Add3ZoneOne(packet)
							}else{
								h265frame = AddBufs(h265frame,Add3ZoneOne(packet))
							}
						}
					}
					SendH265FrameData(rtcDc,h265frame,va.Timestamp.UnixMilli()-start)
			 
				})

				rtcDc.OnMessage(func(msg DataChannelMessage) {
					msg_ := string(msg.Data)
					fmt.Println(msg_)
			 
				})

				rtcDc.OnClose(func() {
					// fmt.Println("hd265 dc close")
					nInSendH265Track--
					// syschan <- struct{}{}
				})
			})
			// frame := VideoDeConf( v.DecoderConfiguration )
			// fmt.Println("*****b*******:",frame)
			// F := (*VideoFrame)(frame)
			go v.Play(suber.IO, func(va *common.AVFrame[common.NALUSlice]) error {
				
			// 	// VideoDeConf(v.DecoderConfiguration).GetAnnexB()
			// 	b: = F.GetAnnexB()

				// frame := (*VideoFrame)(va)
				// fmt.Println("*****frame*******:",frame.GetAnnexB())
				// fmt.Println("*****va.IFrame*******:",va.IFrame)
				// fmt.Println("*****frame.IFrame*******:",frame.IFrame)
				var h265frame []byte

				if va.IFrame { //如果是关键帧,增加关键帧头
					h265frame = []byte{0,0,0,1,64,1,12,1,255,255,1,96,0,0,3,0,176,0,0,3,0,0,3,0,123,172,12,0,0,56,64,0,5,126,66,168,0,0,0,1,66,1,1,1,96,0,0,3,0,176,0,0,3,0,0,3,0,123,160,3,192,128,16,229,141,174,228,203,243,112,16,16,16,64,0,3,132,0,0,87,228,40,64,0,0,0,1,68,1,192,242,176,59,36}
				}

				for _, packets := range va.Raw {
					for _, packet := range packets {
						if len(h265frame)==0 {
							h265frame = Add3ZoneOne(packet)
						}else{
							h265frame = AddBufs(h265frame,Add3ZoneOne(packet))
						}
					}
				}
				// timestamp := time.Now().UnixMilli()
				// fmt.Println("*****h265frame*******:",h265frame)
				SendH265FrameData(rtcDc,h265frame,va.Timestamp.UnixMilli()-start)
				// SendH265FrameData(rtcDc,h265frame,frame.Timestamp.UnixMilli()-start)
				// SendH265FrameData(rtcDc,h265frame,timestamp-start)
				return nil
			})
		}

	case *track.Audio:
		audioMimeType := MimeTypePCMA
		if v.CodecID == codec.CodecID_PCMU {
			audioMimeType = MimeTypePCMU
		}
		if v.CodecID == codec.CodecID_PCMA || v.CodecID == codec.CodecID_PCMU {
			suber.audioTrack, _ = NewTrackLocalStaticRTP(RTPCodecCapability{MimeType: audioMimeType}, "audio", "m7s")
			suber.PeerConnection.AddTrack(suber.audioTrack)
			suber.Subscriber.AddTrack(v) //接受这个track
		}
	case VideoRTP:
		suber.videoTrack.WriteRTP(&v.Packet)
	case AudioRTP:
		suber.audioTrack.WriteRTP(&v.Packet)
	case ISubscriber:
		suber.OnConnectionStateChange(func(pcs PeerConnectionState) {
			suber.Info("Connection State has changed:" + pcs.String())
			switch pcs {
			case PeerConnectionStateConnected:
				suber.Info("Connection State has changed:")
				go suber.PlayRTP()
			case PeerConnectionStateDisconnected, PeerConnectionStateFailed:
				suber.Stop()
				suber.PeerConnection.Close()
			}
		})
	default:
		suber.Subscriber.OnEvent(event)
	}
}
