package cc

import (
	"time"
)

type Env string

const (
	// 配置中心客户端版本
	CCClientVersion = "cc-go-1.0.0-dev"

	// 环境，目前配置中心支持qa, pg, pre, prd环境, dev环境支持文件方式,格式为key=value
	DEV Env = "dev"
	QA  Env = "qa"
	// 故障演练环境
	PG  Env = "pg"
	PRE Env = "pre"
	PRD Env = "prd"

	// 配置中心默认备份路径
	QADefaultBackupDir  = "/data/etc/cc/sdk/%s/"
	PRDDefaultBackupDir = "/data/etc/cc/sdk/%s/"

	// 服务端地址, 生产环境和预发布的地址相同，通过env参数获取不通环境的配置 (可配置？)
	// QAServerAddr  = "10.104.34.242:19091"
	// QAServerAddr  = "172.25.128.248:19091"
	QAServerAddr  = "172.25.20.74:19091"
	PRDServerAddr = "cc-admin-grpc.1sapp.com:19091"
	QACrt         = `-----BEGIN CERTIFICATE-----
MIIDFDCCAfwCCQDbYTAw8BjYmjANBgkqhkiG9w0BAQsFADBMMQswCQYDVQQGEwJH
QjEOMAwGA1UEBwwFQ2hpbmExFDASBgNVBAoMC2dycGMtc2VydmVyMRcwFQYDVQQD
DA5zZXJ2ZXIuZ3JwYy5pbzAeFw0xOTExMjExMTI3NDFaFw0yOTExMTgxMTI3NDFa
MEwxCzAJBgNVBAYTAkdCMQ4wDAYDVQQHDAVDaGluYTEUMBIGA1UECgwLZ3JwYy1z
ZXJ2ZXIxFzAVBgNVBAMMDnNlcnZlci5ncnBjLmlvMIIBIjANBgkqhkiG9w0BAQEF
AAOCAQ8AMIIBCgKCAQEAy3UVweHGaC/X34fLfehmVK7ZSY/wT9Ie0LuoyAaWjPPW
xm/5M8wO2/3Pdca/cFHJRB046SI46/xk/A1E6jr57pUO5CE78xWOQC17bnQmT5Ef
/1fhPEZruVsDqBQVl7Ee0js2vdefNns07UfCCQZVoW1SUO9dcfcSVsBYIy2Xa0g0
jLs+jxHvzsZkSCIfZ15iPc/xfbm82UV1sh3AYHPfhegDYDrYuGzcR1/BSXikT80z
rVq9lxmPFXJDG7VnMJ5X5oxVKpANBODDlPwTvR0i8ftJByOtj4/25cNlGZ1BwwRA
ze/TaaEJarzstAMJ5O/OtodDBky6C+ueK/CB7UsihwIDAQABMA0GCSqGSIb3DQEB
CwUAA4IBAQCAeAggizC/8QMfxceyoO2J2Qj3gpn/BML7B5ZIUbEG+/cgSPFCybeP
XSCuT3Uu/K5c63pKSMCRfOB9jgP5TWjtY5jaDGi4Lj1UkAwiZtiQyTrnQwosGQIA
a6uAGu+72GUfgt4UfUglth3MBtVQWM529yQYEbf7CweCnXagOOf2GULUfp8Z2BQF
i67i954zG58H54uc+grYzx1Bz83XRszvGcCW22Bb2eJwd86Sq4nLFKSYTz33VPTB
QJbZjAQMQKZBhWWO+UfgDkLNO9Y41nPilqun1aTL5sB+B5kYMlNpLRwzAuT5BGzr
hpXWAUuggcNipGWhZakOJ/k8RucVl+SR
-----END CERTIFICATE-----`

	PrdCrt = `-----BEGIN CERTIFICATE-----
MIIDFDCCAfwCCQCmFZgjVLTmyDANBgkqhkiG9w0BAQsFADBMMQswCQYDVQQGEwJH
QjEOMAwGA1UEBwwFQ2hpbmExFDASBgNVBAoMC2dycGMtc2VydmVyMRcwFQYDVQQD
DA5zZXJ2ZXIuZ3JwYy5pbzAeFw0xOTExMjExMTI3MjNaFw0yOTExMTgxMTI3MjNa
MEwxCzAJBgNVBAYTAkdCMQ4wDAYDVQQHDAVDaGluYTEUMBIGA1UECgwLZ3JwYy1z
ZXJ2ZXIxFzAVBgNVBAMMDnNlcnZlci5ncnBjLmlvMIIBIjANBgkqhkiG9w0BAQEF
AAOCAQ8AMIIBCgKCAQEAvhebyV2V8J9QaQQFQUW4BoVXGx4SsE6zpxcN9fN1hUEg
ekzu2jZcBWXur/JtZAuyGyFUq9WCHeayvedpnbu6fDMNvfs1sH81a8q4e1Axkl71
0UAye/d0bC3RKlYMLXjGRvKsX0WQ7ouuF3Wu6hoVFKcdv4L1f0R9bueZao5i0AJF
RGgSzadTheWdsU/eY+T8Y/DfjJcsZRgeK/c5wuIfadTGVZAR4/9+1+vIXTM4ZAhQ
6R/WsJ687iezdCGQ+dXTwP9qH1SV/ul/+1EbaKPpSIyCq264YNYzzVDar60I9pBw
FA73UHr6I95ZBuqr5A04phv/JuH97xtJG61hzpd/HQIDAQABMA0GCSqGSIb3DQEB
CwUAA4IBAQArb3tTeUJq84Vj3v1otZZuS46Xepgip+DwWmbXyDUUtxhgLLBnaQow
9f8r2mAKmDnbhdWOVtNf718XYYwR3xVNEfVAV5az/dWDrou7H5uAhd8jREc68vnx
X873xbZOxUieEHQZjr3Y8NS/m7ZisSR6Q3Kdn8J4GfKuKCd8VPV8Ko1PnliNUbRY
PgJ99u5ug3JTF62FS+OyaPC7y1TcOm0nU7LcgEkHIJCwa7rNHOY8U8mRCvNr87ND
ZCf76URR2lgduWkAnexMVDRB2HLZnWWtmhRjiJadR6v/6eBzUj5FJE2/hcxsmizP
Jc2pTbBdTTpqPkUzE/DCwdI6beKoEs7i
-----END CERTIFICATE-----`
	// gRPC拉取超时时间
	GRPCTimeout = 10 * time.Second

	// backoff
	Jitter                 = 1 * time.Second
	DefaultJitterFraction  = 0.10
	DefaultInitialInterval = 500 * time.Millisecond
	DefaultMultiplier      = 2
	DefaultMaxInterval     = 60 * time.Second

	// heartbeat interval
	HeartbeatInterval = 10 * time.Second
	// ticker update interval
	TickerUpdateInterval = 1 * time.Minute
	// gRPC 发送接收超时时间
	gRPCSendTimeout = 5 * time.Second
	// debug
	Debug = false
	// callback timeout
	CallbackTimeout = 5 * time.Second
)
