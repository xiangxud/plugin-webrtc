# webrtc 插件

提供通过网页发布视频到monibuca，以及从monibuca拉流通过webrtc进行播放的功能,该分支增加了对h265的视频流播放,
该插件为项目:
[https://github.com/langhuihui/monibuca.git](https://github.com/langhuihui/monibuca.git)
v3版本组件,后续会添加v4版本的h265播放


# 基本原理

通过浏览器和monibuca交换sdp信息，然后读取rtp包或者发送rtp的方式进行

# API
- /api/webrtc/play?streamPath=live/rtc
用于播放live/rtc的流，需要在请求的body中放入sdp的json数据，这个接口会返回服务端的sdp数据
- /api/webrtc/publish?streamPath=live/rtc
同上
