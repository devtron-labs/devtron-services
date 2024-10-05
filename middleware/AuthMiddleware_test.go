/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package middleware

import (
	"fmt"
	"github.com/devtron-labs/authenticator/oidc"
	"github.com/golang-jwt/jwt/v4"
	"log"
	"net/http"
	"strings"
	"testing"
	"time"
)

type TestClaims struct {
	Groups map[string]string `json:"groups"`
	jwt.RegisteredClaims
}

func Test_ReconstructSplitToken(t *testing.T) {
	verifySplitAndJoinToken := func(token string, tt *testing.T) {
		log.Print("token len : ", len(token))
		cookies, err := oidc.MakeCookieMetadata(oidc.AuthCookieName, token)
		if err != nil {
			t.Error(err)
		}
		r, _ := http.NewRequest("GET", "https://devtron.ai", nil)
		for _, c := range cookies {
			keyVal := strings.Split(c, "=")
			cookie := &http.Cookie{
				Name:  keyVal[0],
				Value: keyVal[1],
			}
			r.AddCookie(cookie)
		}
		finalToken, err := oidc.JoinCookies(oidc.AuthCookieName, r.Cookies())
		if err != nil {
			t.Error(err)
		}
		if token != finalToken {
			tt.Fail()
		}

	}

	generateJWTToken := func(numClaims int) string {
		mapClaims := make(map[string]string)
		secret := []byte("T0p53cr3t")
		for i := 0; i < numClaims; i++ {
			mapClaims[fmt.Sprintf("key-%d", i)] = time.Now().String()
		}
		unsignedToken := jwt.NewWithClaims(jwt.SigningMethodHS256, TestClaims{Groups: mapClaims,
			RegisteredClaims: jwt.RegisteredClaims{
				Issuer: ApiTokenClaimIssuer,
			}})
		token, _ := unsignedToken.SignedString(secret)
		return token
	}

	t.Run(">5kb token", func(tt *testing.T) {
		token := "eyJhbGciOiJSUzI1NiIsImtpZCI6Ijk2NDlhYzxxxyyyzzziOTNiNjBjMGRiMWVlN2Y1MmQ1ODZjOTMyNzAifQ.eyxxxyyyzzzodHRwczovL2RldnRyb24uazguZGVzYXJyb2xsby5lbXQuZXMvb3JjaGVzdHJhdG9yL2FwaS9kZXgiLCJzdWIiOiJDaVJrTjJVMk1tUTBZeTFpWmpObExUUTVaR010T1dNeFppMW1ZVFF3T0RZNE16WmpNeklTQ1cxcFkzSnZjMjltZEEiLCJhdWQiOiJhcmdvLWNkIiwiZXhwIjoxNzEyMzA1NDk1LCJpYXQiOjE3MTIyMTkwOTUsImF0X2hhc2giOiI1czViaUFRNGVUNXAzWVBFWElzOGFRIiwiY19oYXNoIjoiMVZfVnRnakxCS3pjZHNRMFVQSHotdyIsImVtYWlsIjoiYW5yZWNpb0BlbXRtYWRyaWQuZXMiLCJlbWFpbF92ZXJpZmllZCI6dHJ1ZSwiZ3JvdXBzIjpbIlJlZGlzZcOxbyBkZSBQcm9jZWRpbWllbnRvcyBGdW5jaW9uYWxlcyIsIlZhbG9yYWNpw7NuIG9mZXJ0YXMgQnVzIGEgZGVtYW5kYSIsIkZvbmRvcyBFdXJvcGVvcyAoaW50ZXJubyBEaXIuIFRlY25vbG9nw61hKSIsImxpc3RhLlBIUi5wcm9jZXNvcy5HRVgiLCJMaXN0YS5wZXRpY2lvbmVzLmFwbG5lZ29jaW9fYWdlbmNpYSIsIlByb3llY3RvcyBldXJvcGVvcywgaW50ZXJuYWNpb25hbGVzIHkgb3RyYXMgY29sYWJvcmFjaW9uZXMiLCJMaXN0YS5wbGFuaWxsYVNhbGlkYSIsIkxpc3RhLnBldGljaW9uZXMuYXBsbmVnb2Npb19nZXMtaW5jaWRlbmNpYXMiLCJPcGVyYXRpb25zIiwiTWFycXVlc2luYXMgNC4wIiwiQ2FyZGlvTUFEIChQdWVzdGEgZW4gTWFyY2hhKSIsIlZpc29yIGRlIFZpZGVvV2FsbCIsIkxBTlpBTUlFTlRPUyBZIFBSRVNFTlRBQ0lPTkVTIiwiQmljaVBBUksiLCJEZW1hbmRhIGVuIFRpZW1wbyBSZWFsIiwiVmlzdWFsaXphY2lvbmVzIFBvd2VyQkkiLCJTbWFydCBCdXMgTWFkcmlkIC0gQXV0b2LDunMgYSBsYSBkZW1hbmRhIiwiQXBhcmNhbWllbnRvcyBEaXN1YXNvcmlvcyIsIkxlbmd1YWplIE5hdHVyYWwgeSBCb3RzIiwiTUFEUklEIE1PQklMSVRZIDM2MCIsIkxpc3RhLkNhbGlkYWQiLCJsaXN0YS5pbnRlZ3JpYS5hcHAubmVnb2NpbyIsImxpc3RhLnVzdWFyaW9zLmRlc2Fycm9sbG8iLCJsaXN0YS5wYW5kb3JhLkdvb2dsZVRyYW5zaXQiLCJBY3VlcmRvcyBDb2xhYm9yYWNpw7NuIEJpY2lNYWQiLCJMaXN0YS5wZXRpY2lvbmVzLmFwbG5lZ29jaW9fc2llIiwiUGFzYXJlbGEgZGUgUGFnb3MgRU1UIiwiTVBBU1MiLCJsaXN0YS5pbmNpZGVuY2lhcy5iaWNpbWFkIiwiRm9ybWFjacOzbiBIVE1MK0NTUytKUyIsIlByb3RlY2Npw7NuIGRlIERhdG9zIGVuIEVNVCBNYWRyaWQiLCJGb3JtYWNpw7NuIiwiTWVkaW8gQW1iaWVudGUiLCJTb2NpYWwgTWVkaWEiLCJHb2JpZXJubyBkZWwgRGF0byAoRU1UIC0gRGVzaWRlZGF0dW0pIiwiVGVjbm9sb2fDrWFzIGRlIE5lZ29jaW8iLCJHZXN0acOzbiBkZSAgbGEgRXhwbG90YWNpw7NuIiwiTVdhbGxldCB5IFNESyAgSU5FVFVNK1NPTFVTT0ZUK0VNVCIsIkVzdHVkaW8gVHJhemFiaWxpZGFkIENsaWVudGVzIiwiTXkgIFNlbGYgVGVhbSIsImxpc3RhLmluY2lkZW5jaWFzLlNQTyIsIlNBRU5leHQiLCJMaXN0YS5yZXN1bWVuQklUIiwibGlzdGEucGFuZG9yYS5FbGV2b24iLCJHcsO6YXMgXyBMaXF1aWRhY2nDs24gYXV0b23DoXRpY2EiLCJsaXN0YS5wYW5kb3JhLmVudmlvcy5DUlRNIiwiZUNvbW1lcmNlIE1wYXkiLCJ1QXp1cmVfQWRvYmVTaWduIiwiVGVjbm9sb2fDrWEgeSBTaXN0ZW1hcyBkZSBJbmZvcm1hY2nDs24iLCJCaWNpQk9YIiwiRWxlY3Ryby1FTVQiLCJBcGFyY2FiaWNpcyIsIkV4cGVyaWVuY2lhIGRlIFVzdWFyaW8iLCJtdVNlbmRBc19tb2JpbGl0eWxhYnMiLCJSZW5vdmFjacOzbiBCaWNpTWFkIEludGVncmFsIiwiTGlzdGEuSmVmZXMuU2VydmljaW8iLCJNYWFzNEFsbCIsIlVzdWFyaW9zX0FjY2Vzb19FeHRlcm5vXzM2NSIsIkNPTUlUw4kgU0VHVVJJREFEIERFIExBIElORk9STUFDScOTTiBZIFBST1RFQ0NJw5NOIERFIERBVE9TIiwiQnVzUmFwaWRUcmFuc2l0IiwiTGlzdGEuR0lTIiwidU9mZmljZTM2NV9Hcm91cHNfTWFuYWdlcnMiLCJMaXN0YS5wZXRpY2lvbmVzLmFwbG5lZ29jaW9fZ2V4Y29uIiwiU2VndWltaWVudG8gVGFyZWFzIG90cmFzIERpcmVjY2lvbmVzIEVNVCIsInVBenVyZV9EZXZ0cm9uIiwiRHJlYW0gVGVhbSIsIkVzdHJhdGVnaWFzIGUgaW5pY2lhdGl2YXMgZGUgRGF0b3MiLCJTb2x1c29mdC1FTVQgVGFyZWFzIHkgRG9jdW1lbnRhY2nDs24iLCJJbmNpZGVuY2lhcyBCaWNpTWFkIiwiTW9iaWxpdHkgTWFkcmlkIiwiTW9kZWxvIFByZWRpY3Rpdm8gT2N1cGFjacOzbiBkZWwgQlVTIiwibGlzdGEucmVzcG9uc2FibGVzLmRlcGFydGFtZW50byIsIkNvbnZvY2F0b3JpYXMgRm9uZG9zIHkgQXl1ZGFzIChOZXh0R2VuLCBGRURFUi4uLikiLCJMaXN0YS5wZXRpY2lvbmVzLmFwbG5lZ29jaW9fY3VhZHJvc2JkciIsIkNvbmN1cnNvIGRlIGlkZWFzIGRpc2XDsW8gbnVldmEgYXBwIiwibGlzdGEuZGlyZWNjaW9uLnRlY25vbG9naWEiLCJTaXN0ZW1hIGRlIEluZGljYWRvcmVzIiwiTGlzdGEuaW5mb3JtZXMuaW5jaWRlbmNpYXMuQklUIiwibGlzdGEucGFuZG9yYS5zZXJ2aWNpb3Mud2ViIiwibGlzdGEuYXl1ZGFudGVzLnRlY25pY29zIiwiT3JnYW5pemFjacOzbiAtIERpcmVjY2nDs24gZGUgVGVjbm9sb2fDrWEgZSBJbm5vdmFjacOzbiIsIlBsYW5pZmljYWNpw7NuIHkgQ3VhZHJvcyIsIkF1dG9iw7pzIGEgZGVtYW5kYSAtIFdvbmRvIFx1MDAyNiBFTVQiLCJMaXN0YS5wZXRpY2lvbmVzLmFwbG5lZ29jaW9fZ2V4YnciLCJHSVMgLSBTaXN0ZW1hIGRlIEluZm9ybWFjacOzbiBHZW9ncsOhZmljYSIsIkxpc3RhLkNsYXNpZmljYWNpb25CSVQiLCJWaXNvciArIEtpYmFuYSAvLyBCaWNpTUFEIiwibGlzdGEuZ2VzdGlvbi5kZXNhcnJvbGxvLnNvZnR3YXJlIiwiQmljaU1BRCIsIkNlbnRyb3MgZGUgT3BlcmFjaW9uZXMiLCJDQVJESU9NQURfVGVjbm9sb2fDrWEiLCJsaXN0YS5zaXN0ZW1hcy5FTVQiLCJBdXRvYnVzZXMgeSBUZWNub2xvZ8OtYSIsIkJQTSBTZWd1aW1pZW50byBIYWJiZ" +
			"XIrRU1UIiwiSG9qYSBkZSBydXRhIGVsZWN0csOzbmljYSIsInVBenVyZV9Hb29kSGFiaXR6IiwiQXBsaWNhY2lvbmVzIGRlIE5lZ29jaW8iLCJsaXN0YS5heXVkYW50ZXMudGVjbmljb3MucHJpbmNpcGFsZXMiLCJMaXN0YS5zZGEiLCJUYXJqZXRhIGRlIEVtcGxlYWRvIiwiR2VzdG9yIGRlIENvbnRlbmlkb3MgeSBDYW1wYcOxYXMgcGFyYSBBcHBzIiwiTGlzdGEudG9kb3MiLCJHZXN0acOzbiBkZSBsYSBFeHBsb3RhY2nDs24iLCJMaXN0YS5wZXRpY2lvbmVzLmFwbG5lZ29jaW9fc21zIiwibGlzdGEuYXBsaWNhY2lvbmVzLm5lZ29jaW8iLCJsaXN0YS5kaXZpc2lvbi5zaXN0ZW1hcyIsInVPZmZpY2UzNjVfRTNfVklQIiwiRGlyZWNjacOzbiBkZSBUZWNub2xvZ8OtYSBlIElubm92YWNpw7NuIiwiTmV3IEJpY2lNQUQiLCJsaXN0YS5hdmlzb3MuR0VYQ09OIiwiTWlncmFjacOzbiBNTSBFTVQiLCJsaXN0YS5leGNsYWltZXIucGFjaWZpY28iLCJQcm95ZWN0byBTUE8iLCJJbnRlZ3JhY2lvbmVzIGVuIEFwYXJjYW1pZW50b3MiLCJUcmFzbGFkbyBkZSBMYSBFbGlwYSBhIEZ1ZW5jYXJyYWwiLCJFdm9sdWNpw7NuIGVjb3Npc3RlbWEgTUIzNjAiLCJEZXNpZ24gSWRlYXMiLCJSZWR1Y2Npw7NuIGRlbCBTZXJ2aWNpbyIsIkVzdGFiaWxpemFjacOzbiBiaWNpbWFkIiwiU2VndWltaWVudG8gdGVtYXMgQ1JUTSIsImxpc3RhLmNhbGlkYWQuYWlyZSIsIkRldmVsb3BtZW50IiwiUHJvdG9jb2xvIGRlIFBhbmVsZXMgRXh0ZXJpb3JlcyIsIk51ZXZhIENvbnNvbGEgIGRlIENvbmR1Y3RvciIsImxpc3RhLkJXIiwiQ29tcGV0aXRpdmUgSW50ZWxsaWdlbmNlIiwiRGlzZcOxbyBkZSBudWV2YSBDb25zb2xhIGRlIENvbmR1Y3RvciIsImxpc3RhLlBIUi5wcm9jZXNvcy5jdWFkcm9zIiwibXVTZW5kQXNfb3BlbmRhdGEiLCJCaWNpbWFkLU1QYXNzIiwiRnJlZS1mbG9hdGluZyBBdmFuemEgQmlrZSAtIEVNVCJdLCJuYW1lIjoiQW5kcsOpcyBSZWNpbyBNYXJ0w61uIn0.nlp-RKisxYe24vK44k18Eqi44XKIyk0G6cYc5YTmP4B6AD1eV0vMU9YCzEcmTpgp4t1LroYX9Kjox6tOgY1EY6XbbCxJpJ8w9aXP-mMXH5BaiHP1nZVKCWCqwaKxQAwTq9qI30-NedwsPqNOC3zd7xQPKvt3leBv59mdROVV47jfiX2BptJ5vD2qC-jk9A47FngzzNrvForIqgmE2svUgslGsnE7ywx3D28UHKFhDrD-rHyIeOHXDwgCokMILDez9-P-9k9GYJVoBanHSblYzQZjKIjtwOO_9obW9iVb1t6GNk_8co8YaYUHL8TN4g95_UBI6uWHN5CK-xxxyyyzzz"
		verifySplitAndJoinToken(token, tt)
	})

	t.Run("5 key vals in claims", func(tt *testing.T) {
		token := generateJWTToken(5)
		verifySplitAndJoinToken(token, tt)
	})

	t.Run("10 key vals in claims", func(tt *testing.T) {
		token := generateJWTToken(10)
		verifySplitAndJoinToken(token, tt)
	})

	t.Run("20 key vals in claims", func(tt *testing.T) {
		token := generateJWTToken(20)
		verifySplitAndJoinToken(token, tt)
	})

	t.Run("100 key vals in claims", func(tt *testing.T) {
		token := generateJWTToken(100)
		verifySplitAndJoinToken(token, tt)
	})

}
