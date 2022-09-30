package webrtc

import (
	"io/ioutil"
	"net/http"
	"encoding/json"
	. "github.com/pion/webrtc/v3"
	// . "m7s.live/engine/v4"
)

func (conf *WebRTCConfig) PlayV3_(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Max-Age", "86400")
	w.Header().Set("Access-Control-Allow-Methods", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
	w.Header().Set("Access-Control-Expose-Headers", "*")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Content-Type", "application/json")
	streamPath := r.URL.Path[len("/webrtc/playv3/"):]
	bytes, err := ioutil.ReadAll(r.Body)
	var suber WebRTCSubscriberPro
	// var suber WebRTCSubscriber
	var offer SessionDescription
	if err = json.Unmarshal(bytes, &offer); err != nil {
		return
	}
	suber.SDP = offer.SDP
	if suber.PeerConnection, err = conf.api.NewPeerConnection(Configuration{}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	suber.OnICECandidate(func(ice *ICECandidate) {
		if ice != nil {
			suber.Info(ice.ToJSON().Candidate)
		}
	})
	if err = suber.SetRemoteDescription(SessionDescription{Type: SDPTypeOffer, SDP: suber.SDP}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err = plugin.Subscribe(streamPath, &suber); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if sdp, err := suber.GetAnswerV3(); err == nil {
		ret := WebRtcReturn{}
		json.Unmarshal(sdp, &ret)
		ret.IsH265 = suber.isH265
		byt, _ := json.Marshal(ret)
		w.Write(byt)
		// w.Write(sdp)
	} else {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}


func (IO *WebRTCIO) GetAnswerV3() ([]byte, error) {
	// Sets the LocalDescription, and starts our UDP listeners
	answer, err := IO.CreateAnswer(nil)
	if err != nil {
		return nil, err
	}
	gatherComplete := GatheringCompletePromise(IO.PeerConnection)
	if err := IO.SetLocalDescription(answer); err != nil {
		return nil, err
	}
	<-gatherComplete

	if bytes, err := json.Marshal(IO.LocalDescription()); err != nil {
		return bytes, err
	} else {
		return bytes, nil
	}
}


