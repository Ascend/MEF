// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

package edgeproxy

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/gorilla/websocket"
	"github.com/smartystreets/goconvey/convey"

	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/x509/certutils"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/util"
	"edge-installer/pkg/edge-main/common/configpara"
)

func TestStartEdgeProxy(t *testing.T) {
	p := gomonkey.ApplyFunc(getCertInfo, mocGetCertInfo)
	p2 := gomonkey.ApplyFuncReturn(configpara.GetNetType, constants.FDWithOM, nil)
	time.Sleep(1)
	defer p.Reset()
	defer p2.Reset()
	convey.Convey("Start the entire EdgeProxy project.\n", t, func() {
		err := StartEdgeProxy()
		convey.So(err, convey.ShouldBeNil)
	})
}

func mocGetCertInfo() (*certutils.TlsCertInfo, error) {
	rootCaPem := getRootCa()
	certPem := getCerPem()
	keyPem := getKeyPem()
	kmcData, err := getKmcContent()
	if err != nil {
		return nil, err
	}
	tempCertDir := os.TempDir()
	caPath := filepath.Join(tempCertDir, constants.RootCaName)
	certPath := filepath.Join(tempCertDir, constants.ServerCertName)
	keyPath := filepath.Join(tempCertDir, constants.ServerKeyName)
	kmcMasterPath := filepath.Join(tempCertDir, "master.ks")
	kmcBackupPath := filepath.Join(tempCertDir, "backup.ks")
	if err := fileutils.WriteData(caPath, rootCaPem); err != nil {
		return nil, err
	}
	if err := fileutils.WriteData(certPath, certPem); err != nil {
		return nil, err
	}
	if err := fileutils.WriteData(keyPath, keyPem); err != nil {
		return nil, err
	}
	if err := fileutils.WriteData(kmcMasterPath, kmcData); err != nil {
		return nil, err
	}
	if err := fileutils.WriteData(kmcBackupPath, kmcData); err != nil {
		return nil, err
	}
	kmcCfg, err := util.GetKmcConfig(tempCertDir)
	if err != nil {
		return nil, err
	}
	return &certutils.TlsCertInfo{
		RootCaPath: caPath,
		KeyPath:    keyPath,
		CertPath:   certPath,
		KmcCfg:     kmcCfg,
		SvrFlag:    true,
	}, nil
}

func TestHandleConnectReq(t *testing.T) {
	convey.Convey("Check whether the handleConnectReq message can be properly processed.\n", t, func() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handleConnectReq(w, r)
		}))
		defer server.Close()

		u := "ws" + strings.TrimPrefix(server.URL, "http") + constants.EdgeOmSvcUrl
		ws, _, err := websocket.DefaultDialer.Dial(u, nil)
		if err != nil {
			t.Fatalf("Failed to dial WebSocket server: %v", err)
		}
		defer func(ws *websocket.Conn) {
			err := ws.Close()
			if err != nil {
				panic(err)
			}
		}(ws)
	})
}

func TestGetConnection(t *testing.T) {
	convey.Convey("Simulate a WebSock upgrade failure.\n", t, func() {
		req := httptest.NewRequest("GET", "ws://localhost:8080", nil)
		req.RemoteAddr = "127.0.0.1:8080"
		w := httptest.NewRecorder()
		req.Header.Add("Connection", "upgrade")
		req.Header.Add("Upgrade", "websocket")
		conn, _ := getConnection(w, req)

		convey.So(conn, convey.ShouldBeNil)
	})

}

func getRootCa() []byte {
	return []byte(`
-----BEGIN CERTIFICATE-----
MIIEsjCCAxqgAwIBAgIUH5CaZqxN7uAwOTBKxSMm0IPPkrUwDQYJKoZIhvcNAQEL
BQAwRDELMAkGA1UEBhMCQ04xDzANBgNVBAoTBkh1YXdlaTEPMA0GA1UECxMGQXNj
ZW5kMRMwEQYDVQQDDApodWJfY2xpZW50MB4XDTIzMDgyODA3NTgxMloXDTMzMDgy
ODA3NTgxMlowRDELMAkGA1UEBhMCQ04xDzANBgNVBAoTBkh1YXdlaTEPMA0GA1UE
CxMGQXNjZW5kMRMwEQYDVQQDDApodWJfY2xpZW50MIIBojANBgkqhkiG9w0BAQEF
AAOCAY8AMIIBigKCAYEAr3b67bhXV+BjO/LnfQQbWqTdE31HFbjPMt8n2bVlW3w4
BN6UBecwgPLys0m9Vt53qAjM2Awz7o9b1VNHqmGQu+b99BnjnQhq/VlSYdjkcYRW
WREFvI3o0QflG1JAaGER8ubXtSrjrICAIR6lyHAdIIDIYZyOsbx5SNV1KjuJZmYm
ouFCNFXRrnQ6HoJkvlb/xv4OBTMUUZRPUZf18ocPgYMw8OCFBqr1vkHmhroKRMn4
ILK+0Ao/0crj1vY/NRFLNTPO+O8UCnH6YDC7PnVvFw4Af846qdXtREd/7HaWpmQB
2/2D8q3VsG582SWJrr3KZKi8mIu71Xk3pjZmTrayGkoJQ+M6VdFTSL/iLMWzX1YK
FPm/q7XiXDx19DZClWN8TCpGGRyBxA0H9x7Ej1eEE743nX/HrVqsa8iy1aK0DUG+
3hS42gBodZ/ppxvgUQuY/aTgL+zGBKTrMr/zAHCyotRLL2I2r9df6BOSOc/VX9mP
c9pjhDa8Gqk9SRUVt9UxAgMBAAGjgZswgZgwDgYDVR0PAQH/BAQDAgKEMB0GA1Ud
JQQWMBQGCCsGAQUFBwMCBggrBgEFBQcDATAPBgNVHRMBAf8EBTADAQH/MCkGA1Ud
DgQiBCAUunhjwfWe1P/vUcI2NvxgIYckC2QGfNd5ks2fA7DTcjArBgNVHSMEJDAi
gCAUunhjwfWe1P/vUcI2NvxgIYckC2QGfNd5ks2fA7DTcjANBgkqhkiG9w0BAQsF
AAOCAYEAEzLMUNOlAHK8peL/obKh1PnGjpMrFkQihPe+ijinNvJr7w6tkeHcL5JX
vGe+pUs9EiUqOIMGd8gEBttjFHZGSMPSMzWeINWn++fYLgSOFV+VIRxoRBSoyWA7
BcFnEXaYbSBkri4dq/M+AjQIMqo7Tn2qLAYPiciDeN2oXehJUgmC9maAgT8F+eNV
cD+lXsXTGE67v/ad4q5xg7KgjzipiVSY5iqkQ0vgUXkGp6qY3bWuzFM9PS1qeHJS
/Amt+YR1UFvxFcKuE+nc9kAYuDf4QFmrqnHZ6zDTHO0yeVCbKt2flql6Yagn9xSV
rjr/aQOxZS+M4/IyFhi4h8FKimHP0ZFhu17+NG2p6pFoUmo8V2J+Z7VjMZ2XBle/
w18kc2kmf71iT/+N2C2hBWl0MjExQkrv3zTRPiNC4z/bWZBfYCias//bb4Tdizve
le9MDeFAW3uVsC6lQjAO1MeAulkxWjHyWjk0XWuYgMlWnnpSVZwX9pJg8aDg8Qa5
JowhTAZy
-----END CERTIFICATE-----
`)
}
func getCerPem() []byte {
	return []byte(`
-----BEGIN CERTIFICATE-----
MIIEtDCCAxygAwIBAgIVAPGoPFQkz3KDVONBr0BfWykT2OJkMA0GCSqGSIb3DQEB
CwUAMEExCzAJBgNVBAYTAkNOMQ8wDQYDVQQKEwZIdWF3ZWkxDzANBgNVBAsTBkFz
Y2VuZDEQMA4GA1UEAwwHaHViX3N2cjAeFw0yMzA4MjgxNDA5NDZaFw0zMzA4Mjgx
NDA5NDZaMEgxCzAJBgNVBAYTAkNOMQ8wDQYDVQQKEwZIdWF3ZWkxDzANBgNVBAsT
BkFzY2VuZDEXMBUGA1UEAwwOaHViX3N2cl9zZXJ2ZXIwggGiMA0GCSqGSIb3DQEB
AQUAA4IBjwAwggGKAoIBgQCo3lLFHIDkjBjvkuL5zyKmvfPcliK2Q+Nk7gMK09nx
NE3nVfr/CwS5zVdX46JvvVK0HWc4csIXSJMPee+QMYR9EYXBd2sUxtbcCDRUIfgn
Y/jIzCDyFzG8NEUyfb0wTLeLlWrdS5BihNFOBoEsWkRBfmMT/5hQgnxmcnqCPMkn
35MKvPapZyoEH+m8Eswh54W9RJ9r65UgGXru1vHOFYIiqlGL9uDlM2h3rOMjQ3EU
SeSPlo6rMP6y6F0MtQ0Jp6PPXzgZbPX72t3PUXdP+z7ZC4+Jn6js4inUa3j89a/s
gIkjF1KQkSVwWlcKR4yVsyZV2a0tf/U1QU7keztZPxJ66j1DgD+hm9VCoEPeksmr
RE7yGkh87GAO8dKRjaMsCEXKzjOuR4KG6QdKB8fwhYT2kiAC3CPVvkuPG5/hbuz9
WQw2GHZUg2OPxXFf+SXgs8qVdZRWjSnizEGPYDGgAroEGlMwnd8jocY/L/r/peVr
MHwKD0L+ZXxI6cLpb04LdLsCAwEAAaOBmzCBmDAOBgNVHQ8BAf8EBAMCB4AwHQYD
VR0lBBYwFAYIKwYBBQUHAwIGCCsGAQUFBwMBMCkGA1UdDgQiBCDqpDdQLp2TwpOQ
LOnxf6AyaytU/N70KlqFqZnoPkzsLDArBgNVHSMEJDAigCCZlSaNKI14+PVIN1H2
gkLVkxHLmLZ+gEXjcp+bwXwgWDAPBgNVHREECDAGhwQzJkJFMA0GCSqGSIb3DQEB
CwUAA4IBgQCLrLWKvn0yB5DF2VAc8X0TFZAbfMJh/TRU1oSLvOgwjyNe/cBhf8LT
w6TXi3rjP2wLhgK3ezx/dt/2n47Gys7T+F0tg2ExGQdQUgigQE7Y/TCc7XqtZYmH
eA2lBJdUAX4I+OZo9MTKbQ3c2hNfCtM2NtJb4lFBrXSolHtttEl2RTCyJAVZUYtO
3uSvcsutsPDleyFO5c4kH7xdJ+a2TrYuXyftB6DuwaeAU+2tALYEtM9yLxr5SiGZ
BknqiBgwxw5A1xS93pqnVkBcqumrVJQeXrq4wpZwQh7T2Bi1aS9uYbVogR+ZIcFE
BmewTPp3nlp9Icb3Va5XI9SEdz4BjQbeBIi+CzXpNLQ3U2mgyskukCB3s/+Vrx11
SOUhgKjkRq08K+k9mO6CwRAOgr+eLdvJ2K0/85YLn+VqYvoIV8mEXXpxNctiKHia
92Qn2l+LlmCGdl9JnRdiZ+AxABornQka4Ew7N7s0W0HXcRvpio1gaRbwhFm0sVxr
Ezv20LqYp4I=
-----END CERTIFICATE-----
`)
}

func getKeyPem() []byte {
	return []byte(`AAAAAgAAAAAAAAAAAAAAAQAAAAkok6tkheCIKbZz5ik5upQ0BOll2/+ev4K1xhSrAAAAAAEAJxAAAAAAAAAJpx0jj/vyT` +
		`qC4U8aPbHkSTNLvnNNKJshcWxYmugObpxkSFXEgSDgkViBdKu1U1Ya9MJbY87cE4Z9rVVmRSDJUkR1tj9IHayBh/HOqwff26a066sd2n` +
		`4eP/Ldq+TQ1JjJsuyEx0+cVN5b6O2JUSdnoV7AXgVcdxmZp+s+BMbDMeij6OjwX0Yls82JxY1KInCBtEFZUERidXq/ZyvBs5a77fBTJU6` +
		`iPWiJz3EujA2ZjbMPFroyrfNSQlXBZKAXMGiV+GsDRUNUG1CJh2j7krTtmTO0W2zqjXTBUV9ExpgNgwOSUTg7QBXy3uFrZ05HDLf1pXF6` +
		`pM4rV2hXbcE3hm0hcRwr+WDaqMOEfsBrbheG99/QpnNTC922veoWlnVx4Jb3mdm9UTEJsHZCPwcg7qktE7q2gbigBSscKT4RHwCjlZeEf` +
		`Gwccf91r3HoD94Y99ZOEq5jvmCrvlPS9kXj9/v1bPGa4r/p7jpyFlZCqXfaGEbHPYEZ2N8SCNvZmde8gRiJd2x70ESt4YBp4CQmum0v97` +
		`FxrVVDr4qGQHRifdOLlBfCoMAB1qpPbLeLs2ynPwB+PZB3Ktxo3g6finuJ2BoGN67xr51iazEk7nUY37W48NhIpiKoKbSjlt9IvrgdDUF` +
		`bWHfG06mr25S/fAeLlPgbJ8KUDMdst54EKBXz6oUHLotFxbbC0kfcWtsjJ5LdLbi7D4LCZ3XtFR6kwx7qtUoJrQvNqMKPEHOtmnzGyNlf` +
		`SfR4+v289IWF9bxG3yvgm8CE+HfXXdHl8WY5MB4SGmesLeVopXOLfu5P/phRs/EQyEJXFeoAvZnXkCdaUIMZyavxUVVep2bGTO9Xlm4UH` +
		`BHAbcU69a50TO5TJrgQH8QnsXrBkB/BkVLDiPrGmcYBDbuhYvIYVspCnHM2AZm+hdvm9BR8sJ4jjthWOY0YAXAM/MXTwUVvzu1sAsmBWXK` +
		`YarMwvcj7lcN0kgkiEKK+Cf3OYH8tSJuARGJW9I7rNHoC1KR2v54I9ibImKwqQs3uge7qP9Jyj5/Y2wLEgC0PclyKPjJ82eYGPALXqIEtr` +
		`NC3oMUMy+BTw+PzsC2bpnQqFkoloWat7Bo7f+Xa3kyDd8RiZoRT9pQOk01+Vg7xMMpDCXrElUO1jvXIYe30y8bM5b/aVrm7DJqXoAoXva4p` +
		`qGkiH45SQPtSg9UQYCrbqUsaH3RWqYIQKb7FDj78DvQb3zwgDqq+I1ytd7hLV96xmQmP9UKQa/8EVUOrX/D8f8NQ50vdFJIU4sryN2fHngL` +
		`JVBb6o8GCsxvNbS+oEwtT9Lu7N/eLj1Dt1pfqOXx5Z/CvOh2r7JPdQ6y66hOSktSYTQcSJ9C2nYzZAcmJe3n91Ey9eAQP8bzFqcuvxcTqk1` +
		`TZf/hmjYMsGhP802/P9XjgWavD4NHTgVcTsWPF0W9XskDyCALB1wUw14phvSDYxqBD/MqNJ90PyLmXkcPnQ52RorPaimafDAe+bcLkgfKuR` +
		`RogZdWajpbUkdydEMdl9+L+FEIeIms0MwOY6+P7Q1JCWtZdb6aVGyGfguf4FbR/9/RRtHzPkrX11AQQ/LUZD8iSFmo9IONupT4lljCJmyhA` +
		`skma9B60VqfYYfQk8IK/m39oZ/ooRRuAGQQ7J1snTWVXzmiJdNCAfSmDLhwVOMgIamijZX1oMM9khTA4AqJgIq3h5/RK7hfhfxKNNH4gHgi` +
		`eo0OaPRYU1m/a6kCphJEdXtiwdMPKaLCBROOwCauPEmPIPmRksVePn1B6v6Nl23qcwv5W68NsiML8rW7f+4mty0gs0RACf4tpkvWefaCqQW` +
		`l976uUUaBIEpODX4PQFZMJJ8IF7j1yIVMNqC+NfihZfuXgoIRNWLItCXgEQRhugQkxPeZhYJ0qs1SwJM9e53M+sEY6S6giBnLs1WxicktRA` +
		`DdRI7aQWZMq+cSWK2NjiiWjDdPWmMiDFMvMxGP83QZ2dm3cIT+3mHEwlOWs5BJkbZMQrD5pjzGlLjQsJeRH5la/sK4FnKgfrYXN4XEVktWG` +
		`6YrmEC77IxWl6Xp5QgwQAaK3PC3lt49o4iODI0tGt9MXhoSX9K7KhJNJI1zk5TS4jLXS7Xk96kKgrvyvNxHzJxz1BVyPIRBzyl0N2ZyzU+I` +
		`p6FT6ZLJxoKwRAVmtg+8geiDmXJXrBSXn0njcrTt7UjIotXP37U1cRmYvAXC9ZQgIFFQ4mEQsGJRaqIIWjCLrIP1PB0/YTN36TqFFF5733P` +
		`r7I6jT6c+3tlI1z6MYdWExFfEG++b9SMrdX4MLqriQRSMlGS6VxFQbcX9LybbmHno3sNm1Q4UtuVFGd6NaSPLZ4V/icb6/VT0ni0uG02Jgd` +
		`pD7Ha/FvqLxrAwWeKB9WeKvAeTUM8/DCLzad8ZRHitz9/pFxeJqWq/8r3F9Qt6yzuL9z6eAYiP26FxkiubLOiVt0MZLbDtLUbfKTHFTn9mV` +
		`zWactNzqeD7mT17qD97Nmpd0qYFHRnv4zOZ1/KPdLKr10XAwKgWy2Q823iLG9tyVWLhjJ+p0POcXKeGg3/ukbVuCnKLDaLp7BvOy7FFB4Vn` +
		`1BAAZap9I2z6ewz99M6Gnl5NgiuJREb5KOOIVId5anBJ/XtnmsMEA4pi8q4nmL0ZnorSlm4RPBtOoh+TePTC4sYf/Jubte5bQCSb5cxTT3+` +
		`vnOKzx/K6A7wnDJCah5/y1gQlHMs2WZj3LQmf6sWr12NxMSyDzKJUn3JSX95Nu4pRJhmSLVbGcuRFq375mBrJ+ZjTccUVYQJSBUt+ZjnAkL` +
		`CYebZ8DpMxnseRjwmVoc75/aVyL2I0nrGHatc3sUZ7Y0Ii1lsjJKUnQEDyw9Goqggzkatz6tTCzV4V8lzw3KMBfLIRiltH3Uz5T78Vgq32p` +
		`/JpYn/q1z8DV72qPTiOMjUHViPe2i1/iiuyFfkIR6zYAhTM1acsgXrOqXC8RictgOjx81vIOsyDSGee2oOlkP8FS1En5HQxftjwOjZUwZn` +
		`fhigW8nlk8YpMO0YWqjgKIlNSQVviLrAfXG+adbTGzskyBZzdjYniesbhSPkbHv3IQJbtaCt6IoO8EUCQ3SIuSFjhSPruSuOG7/s5yY+85` +
		`s4KSZWcnWuk4tlGo7+nr9oTJkWvKD+IUPfrs1QOnZvDdj7vXs44rGnVthNk7XQY6KLBmMDVIO3lkiJ8onX+d/k5kZFXHbc6tpuTchGfplx` +
		`SqRjk8tKm3dcPWboqCbnoITyg10gTuMu7q6cNzx34OBN+9KX7DckDPIaeQCTGfuH2SGlHTM01rVYkQ0VM7a7QLAXhj1yI1pr8iT`)
}

func getKmcContent() ([]byte, error) {
	KmcBase64Str := `
X2SXjRlPic+oP47h2wE8DIhCShy3/K1wTkUTpRRGcWwAAgAAB+cIHAA3JAEH8QgZADckBAAAJxBd
RwR/urzgg9inGn6NWCKSM1/WTpMu54y4iHdr20c2fklmj7LfWBZRqXmVdM2zk3xpiEI0/d614ktS
BlGovCvlAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAADP5cuNqIIKXTVIaBqwi1MNn43v
Of2oOSiu6y6ou+bwEwAAAAEAAAABAAAAAAAAAAEBAQEAAAAAAAAAAAAAAAAAAAAAAAAAAAC03EP5
x2ZR7vLQduL4OdAshKaSOb2VhqT/3H/bt/n7ngAAAAAAAAABAAMBAAfnCBwANyQBB+gCGAA3JAYA
AAAHAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAOsI5S6Pc5o+pSIUSJnx
VuEAAAAwAAAAIEFkOH61Yky25CrpxYwVgcZ/dZXqPaQuXqQK4JcSFGubK1urV02PxzBEH+wp0tWA
pAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA
AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAOqeMRsVlv+RmiIcTKLK0dGuZZYNyDJVAMZX24rSkutA=`
	kmcData, err := base64.StdEncoding.DecodeString(KmcBase64Str)
	if err != nil {
		return nil, err
	}
	return kmcData, nil
}
