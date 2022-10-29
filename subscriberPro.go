package webrtc

import (
	"fmt"
	"strings"
	"net"
	"github.com/pion/rtcp"
	. "github.com/pion/webrtc/v3"
	. "m7s.live/engine/v4"
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
					// R := v.DecoderConfiguration.Raw
					annexB := VideoDeConf(v.DecoderConfiguration).GetAnnexB()


					var h265frame []byte

					for _, p :=range annexB {//拼接消息头
						h265frame = AddBufs(h265frame,p)
					}

					for _, packets := range va.Raw {
						for _, packet := range packets {
							h265frame = AddBufs(h265frame,Add3ZoneOne(packet))
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



			go v.PlayFullAnnexB(suber.IO, func(frame net.Buffers) error {
				var h265frame []byte
				for _, packet := range frame {
					if len(h265frame)==0 {
						h265frame = packet
					}else{
						h265frame = AddBufs(h265frame,packet)
					}
				}
				timestamp := time.Now().UnixMilli()
				// fmt.Println("推流:",h265frame)
				SendH265FrameData(rtcDc,h265frame,timestamp-start)
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
