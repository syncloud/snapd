// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2016-2020 Canonical Ltd
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License version 3 as
 * published by the Free Software Foundation.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 */

package sysdb

import (
	"fmt"

	"github.com/snapcore/snapd/asserts"
	"github.com/snapcore/snapd/snapdenv"
)

const (
	encodedCanonicalAccount = `type: account
authority-id: canonical
account-id: canonical
display-name: Canonical
timestamp: 2016-04-01T00:00:00.0Z
username: canonical
validation: certified
sign-key-sha3-384: -CvQKAwRQ5h3Ffn10FILJoEZUXOv6km9FwA80-Rcj-f-6jadQ89VRswHNiEB9Lxk

AcLDXAQAAQoABgUCV7UYzwAKCRDUpVvql9g3IK7uH/4udqNOurx5WYVknzXdwekp0ovHCQJ0iBPw
TSFxEVr9faZSzb7eqJ1WicHsShf97PYS3ClRYAiluFsjRA8Y03kkSVJHjC+sIwGFubsnkmgflt6D
WEmYIl0UBmeaEDS8uY4Xvp9NsLTzNEj2kvzy/52gKaTc1ZSl5RDL9ppMav+0V9iBYpiDPBWH2rJ+
aDSD8Rkyygm0UscfAKyDKH4lrvZ0WkYyi1YVNPrjQ/AtBySh6Q4iJ3LifzKa9woIyAuJET/4/FPY
oirqHAfuvNod36yNQIyNqEc20AvTvZNH0PSsg4rq3DLjIPzv5KbJO9lhsasNJK1OdL6x8Yqrdsbk
ldZp4qkzfjV7VOMQKaadfcZPRaVVeJWOBnBiaukzkhoNlQi1sdCdkBB/AJHZF8QXw6c7vPDcfnCV
1lW7ddQ2p8IsJbT6LzpJu3GW/P4xhNgCjtCJ1AJm9a9RqLwQYgdLZwwDa9iCRtqTbRXBlfy3apps
1VjbQ3h5iCd0hNfwDBnGVm1rhLKHCD1DUdNE43oN2ZlE7XGyh0HFV6vKlpqoW3eoXCIxWu+HBY96
+LSl/jQgCkb0nxYyzEYK4Reb31D0mYw1Nji5W+MIF5E09+DYZoOT0UvR05YMwMEOeSdI/hLWg/5P
k+GDK+/KopMmpd4D1+jjtF7ZvqDpmAV98jJGB2F88RyVb4gcjmFFyTi4Kv6vzz/oLpbm0qrizC0W
HLGDN/ymGA5sHzEgEx7U540vz/q9VX60FKqL2YZr/DcyY9GKX5kCG4sNqIIHbcJneZ4frM99oVDu
7Jv+DIx/Di6D1ULXol2XjxbbJLKHFtHksR97ceaFvcZwTogC61IYUBJCvvMoqdXAWMhEXCr0QfQ5
Xbi31XW2d4/lF/zWlAkRnGTzufIXFni7+nEuOK0SQEzO3/WaRedK1SGOOtTDjB8/3OJeW96AUYK5
oTIynkYkEyHWMNCXALg+WQW6L4/YO7aUjZ97zOWIugd7Xy63aT3r/EHafqaY2nacOhLfkeKZ830b
o/ezjoZQAxbh6ce7JnXRgE9ELxjdAhBTpGjmmmN2sYrJ7zP9bOgly0BnEPXGSQfFA+NNNw1FADx1
MUY8q9DBjmVtgqY+1KGTV5X8KvQCBMODZIf/XJPHdCRAHxMd8COypcwgL2vDIIXpOFbi1J/B0GF+
eklxk9wzBA8AecBMCwCzIRHDNpD1oa2we38bVFrOug6e/VId1k1jYFJjiLyLCDmV8IMYwEllHSXp
LQAdm3xZ7t4WnxYC8YSCk9mXf3CZg59SpmnV5Q5Z6A5Pl7Nc3sj7hcsMBZEsOMPzNC9dPsBnZvjs
WpPUffJzEdhHBFhvYMuD4Vqj6ejUv9l3oTrjQWVC
`

	encodedCanonicalRootAccountKey = `type: account-key
authority-id: canonical
revision: 2
public-key-sha3-384: -CvQKAwRQ5h3Ffn10FILJoEZUXOv6km9FwA80-Rcj-f-6jadQ89VRswHNiEB9Lxk
account-id: canonical
name: root
since: 2016-04-01T00:00:00.0Z
body-length: 1406
sign-key-sha3-384: -CvQKAwRQ5h3Ffn10FILJoEZUXOv6km9FwA80-Rcj-f-6jadQ89VRswHNiEB9Lxk

AcbDTQRWhcGAASAA4Zdo3CVpKmTecjd3VDBiFbZTKKhcG0UV3FXxyGIe2UsdnJIks4NkVYO+qYk0
zW26Svpa5OIOJGO2NcgN9bpCYWZOufO1xTmC7jW/fEtqJpX8Kcq20+X5AarqJ5RBVnGLrlz+ZT99
aHdRZ4YQ2XUZvhbelzWTdK5+2eMSXNrFjO6WwGh9NRekE/NIBNwvULAtJ5nv1KwZaSpZ+klJrstU
EHPhs+NGGm1Aru01FFl3cWUm5Ao8i9y+pFcPoaRatgtpYU8mg9gP594lvyJqjFofXvHPwztmySqf
FVAp4gLLfLvRxbXkOfPUz8guidqvg6r4DUD+kCBjKYoT44PjK6l51MzEL2IEy6jdnFTgjHbaYML8
/5NpuPu8XiSjCpOTeNR+XKzXC2tHRU7j09Xd44vKRhPk0Hc4XsPNBWqfrcbdWmwsFhjfxFDJajOq
hzWVoiRc5opB5socbRjLf+gYtncxe99oC2FDA2FcftlFoyztho0bAzeFer1IHJIMYWxKMESjvJUE
pnMMKpIMYY0QfWEo5hXR0TaT+NxW2Z9Jqclgyw13y5iY72ZparHS66J+C7dxCEOswlw1ypNic6MM
/OzpafIQ10yAT3HeRCJQQOOSSTaold+WpWsQweYCywPcu9S+wCo6CrPzJCCIxOAnXjLYv2ykTJje
pNJ2+GZ1WH2UeJdJ5sR8fpxxRupqHuEKNRZ+2CqLmFC5kHNszoGolLEvGcK4BJciO4KihnKtxrdX
dUJIOPBLktA8XiiHSOmLzs2CFjcvlDuPSpe64HIL5yCxO1/GRux4A1Kht1+DqTrL7DjyIW+vIPro
A1PQwkcAJyScNRxT4bPpUj8geAXWd3n212W+7QVHuQEFezvXC5GbMyR+Xj47FOFcFcSZID1hTZEu
uMD+AxaBHQKwPfBx1arVKE1OhkuKHeSFtZRP8K8l3qj5W0sIxxIW19W8aziu8ZeDMT+nIEJrJvhx
zGEdxwCrp3k2/93oDV7g+nb1ZGfIhtmcrKziijghzPLaYaiM9LggqwTARelk3xSzd8+uk3LPXuVl
fP8/xHApss6sCE3xk4+F3OGbL7HbGuCnoulf795XKLRTy+xU/78piOMNJJQu+G0lMZIO3cZrP6io
MYDa+jDZw4V4fBRWce/FA3Ot1eIDxCq5v+vfKw+HfUlWcjm6VUQIFZYbK+Lzj6mpXn81BugG3d+M
0WNFObXIrUbhnKcYkus3TSJ9M1oMEIMp0WfFGAVTd61u36fdi2e+/xbLN0kbYcFRZwd9CmtEeDZ0
eYx/pvKKaNz/DfUr0piVCRwxuxQ0kVppklHPO4sOTFZUId8KLHg28LbszvupSsHP/nHlW8l5/VK6
4+KxRV2XofsUnwARAQAB

AcLDXAQAAQoABgUCV83kkgAKCRDUpVvql9g3IA9hIADAkn4VXnJIFblhMSBe6hbTy7z6AfOhZxXR
Ds/mHsiWfFT6ifGi9SpZowhRX+ff57YvFCjlBqMYLKYE0NsFQYEUc5uBWiFZwC0ENydNhO23DV1B
elTSs6mr9duPm1eJAozFrQETOD1kz5BIamqBUeaTczjM+9l5i485Ffknbc+EaGOrtMEap0GqjByQ
u+ykZGvryVQ447avgjvFsMtA0quFi+SoW9PT/9D26e5rD7RIICYWG8mzFRn5Isqs/X4W1uAiKQe9
pqHMbdNr/FCWX5ws0/nMaOq+b0z4EIIXIfT0JmIlFDQsAgFVnKwYw+zs32cTw4XuzvMhgMDtCowD
YodhiO/5AOMsMMV0qBsYxbIPJIEz7b6gwTYEJoTVkqTit6o3UgWrAy+p4Y7t0ickYIHgwiuKRS9E
fu0Ue+32NFp0XFqZElfXLK/U2yjto+fJXu6uAELsXesfFGIOp/nbRbNavUt9jAJeO7ftQczgf39T
YfA0OKerP5gAOd4+aO3gATPUjfWPsJ9908XC7QqK2BwS1kh/fMrd95mxcmXdF1bBElszKwaToBVQ
1m52EYp06kkPyOu+fGKFAoIMafcV/2Ztz1WMo/Vp0iP/r0WAtBDw6sDJyWOfRjUEvP7BBdEzraHV
VblbSrKzhYeEGdMDi6kFC+KEzfPDPFJX1l3saPBkz9VDuESbktyObQp9VfkFKYBgBnw3msQJk+6k
G4t0o3/DZ7qz/kTJXMogG26Z/FsMhPERsaLTbWRJ3WRyXX8COaTladSf8bG0Oib19outnjuvpjQ0
qEV9eeGRBlx9mbidSYH95cj0zD2DKpeSZ83M5K1pFg+8RKToGElGTTk8vtdTfDVbmi3+QntfLq+z
ZMgs2+SmCWrV/MPC04Dl00CXywdKPyf6toomqRP7A5fS7W8P9fdPn+a8JCblcleGj9nvJXBQjue7
97rofCEszhKhoE9fMCIUcSoTU9YAm5Jr+qclSEbV1pzwTvZ8auMIXtzEZV5n4aK4WPDV+lYCadrL
DlvJSJRuXRvIMbmvU9b8NxgG8AS88BkX3L9vlOpkMculwG1/iooQvxuFaJDargt370wAQo0lCpG3
MxnsSusymwnYegvvvr7Xp/KBLZK1+8Djzm3fwAryp4qNo29ciVw3O9lFKmmuiIcxSY0bauXaK6kv
pTnYkmx7XGPF7Ahb7Ov0/0FE2Lx3JZXSEKeW+VrCcpYQOY++t67b+jf0AV4rZExcLFJzP6MPMimP
ZCd383NzlzkXK+vAdvTi40HPiM9FYOp6g8JTs5TTdx2/qs/SWFC8AkahIQmH0IpFBJep2JKl2kyr
FZMvASkHA9bR/UuXDvbMzsUmT/xnERZosQaZgFEO
`

	encodedSyncloudAccount = `type: account
authority-id: syncloud
account-id: syncloud
display-name: Syncloud
timestamp: 2016-04-01T00:00:00.0Z
username: syncloud
validation: certified
sign-key-sha3-384: hIedp1AvrWlcDI4uS_qjoFLzjKl5enu4G2FYJpgB3Pj-tUzGlTQBxMBsBmi-tnJR

AcLBUgQAAQoABgUCV6yoQQAAelEQAEdSECpdmV5a2G5VMBzJFuHQUU1FzgZ7gPQjc3l0BibDWm8O
rDi7IT3L80OkqS2AoQgHS5KtEvKqEmhyfcdzcXgvCkHR5kucRBJJaPy8z6gGMhzZIPlc+EqY+Cvb
/MQPLvtYYvtAxq1vWz+aDGGwk2Z/dFUG+wofvNWodz400gYTZeFOCZwStBD84S7iY/3pMQgC3+SO
QMr/VI+bgmOukFqZL0cX4ReiuUs2W45V6EC81UGBjk+k7AVTEXMR1Xo8f0yiRzlLoEdKQMCOC45Q
n4eedjCToGRPFcktM0QhgfbcpPIQKHNqKGGvtQQXvW5PIZ7AS4rTfQScXTn1dqDsL/ZVdasvOpCP
5o4WvoWMoU8+Hm4n6ckw4sXn//PZIQrtnkp2DO+9JXXZasIPg4k1mvUQ5Kb9qCcBbaM+OO1izOoC
3PY8xHNQNfHNHwBMewhnU2NpdTS0mTepN/8iFsDT1vSZ28OE2hgbu1ltqx4AsRkCVyFFx6N6OYm2
UDNozU9K5w0NY4u9HSTDz4KrBIalAaKY72CIUqeVsmAcYatXglbj7dVTZTw75M0v1thQiSoKFqHw
CHykZ6BJRgminY1FqOg7tvqTwzYM7lwaE3K8JpAyzie7v+OSLSxy1vlwUmT2lT+h1i28/w+r+R3Q
C0QC8xuHSvOv3YRtzKna3smAfRlB
`

	encodedSyncloudRootAccountKey = `type: account-key
authority-id: syncloud
public-key-sha3-384: hIedp1AvrWlcDI4uS_qjoFLzjKl5enu4G2FYJpgB3Pj-tUzGlTQBxMBsBmi-tnJR
account-id: syncloud
name: root
since: 2016-04-01T00:00:00.0Z
body-length: 717
sign-key-sha3-384: hIedp1AvrWlcDI4uS_qjoFLzjKl5enu4G2FYJpgB3Pj-tUzGlTQBxMBsBmi-tnJR

AcbBTQRWhcGAARAA8dC6HP+NfM5sNgCHH+bsQv4YLIR8glPfJ+HEXyaYdNO1+oFyX4nx7CpV5Umu
TYs7DPVpToAiN3snpBdPPKu5UEzkQ6OGDucf2bZnAInj7WzKwGnOA/Y/uQMduIyeFZ4mLnUNcF+M
e8LV0aS/pQhEdBUuRxEOi9zlv0p7X1bUs6LIUTubu6+smFtbdBBNOD+0qrvjf7CvsScrTsQswtvw
cLoB4GX94wK6RQrlkmYJPUFZqkdWt7cp0iq8d+Ts8UnT8sgWuFzkMCgBKritS7/545mE8AE0fsyF
Gt5+0jcjgs9LDk5gRO7EgoFLXPsEBdiLdVms7OGAwPGG00wfFYL3ho4PCfKq+mH0kOgUAynlJ7x8
MCR92eWEi/ylHXiO0jnRY8UsutrM76eLN41iUla/6j5DcsXxQB/xzlYkUdtXtYrn6L/DTsnixclu
3ogPzlPEFyVxv0vWIgkKLWXj2JRRt2uqe3K33TvdF0H+m6snZTStn7VY3if9fvyx14+tKh16ucdQ
a1zzJoTKTqYWX9B+ZfENGKJUnhTP0x7Cm6lg3EUGay/b5hsA4DBoqShuf/N0jVLojdhxi3Ck/DBN
lqCD0zy4uzvinjX+b4ay+LKBE3N15AsfEkWIwzI+1OdDlOWWqOxJkM6lrQ5hRQ1fHZoCiGjHbjeE
1RIFO2TAw2tpyUcAEQEAAQ==

AcLBXAQAAQoABgUCV8656QAKCRBMcZp594FxpNWlEADQgBlROdBTHpdZ3/9BbasxenUC3VXusMeK
0DmnsHrsAsyVk6xiHQQ3hWxvXKWoDkDsOhUqcQTsDBcIaZ18+qwpQciyItd+w3d7SSJ+MKSUpwsB
NOdgw1ykj7l1M/W7xAAPscFoV1xVSk9+rsLYFYDe23R+ecyotSmF+4QHj5b+hXeVIOUaqQTl5xPC
h0zVYNIUWv42q4Z+hiBS8+8UJ0G+7z/27XORkGHY6TXCt0aph7s5egr8Lm+/jq7c95HVsa7DwSpv
SqPajRnlyLiHFXUYAUPEU9oDgPwtLsqUkFfrv1WZ3ja1rDexgKBta+8BRyCAq3gPcMAjhiHXdjoW
90p893l9N6K82RiEOO9ic0pEezjQldg97oU+ajXNm3ryns+HX6hRd39rpzIsrbVdbCqun4RwMbCM
EVxgC/cuxMGcS40Co3O8wG3H/WIWOqcRQfolQTexmyzQljYt9WyWJdXmtPtaMzQGbOqE/dIjOK9j
xvrghVU4kX6fJFwPi+azMrluHV+WGSVxPCuLW8o2aipjOd1/bUQCL5OwRuaEWuLCiV01J8H/JjWV
hL4gGVqEM2KEPIDwY2yqX36jE7uN9O+mIPnS4Tdj0JQ5ZD1qh34wv+4QvhgNeyP120nuS1ykO9X0
A806uPC5QK1+cgRMUz8zJ0afDNwE/DvpBQvE5CIi9A==
`
)

var (
	trustedAssertions        []asserts.Assertion
	trustedStagingAssertions []asserts.Assertion
	trustedExtraAssertions   []asserts.Assertion
)

func init() {
	canonicalAccount, err := asserts.Decode([]byte(encodedCanonicalAccount))
	if err != nil {
		panic(fmt.Sprintf("cannot decode trusted assertion: %v", err))
	}
	canonicalRootAccountKey, err := asserts.Decode([]byte(encodedCanonicalRootAccountKey))
	if err != nil {
		panic(fmt.Sprintf("cannot decode trusted assertion: %v", err))
	}
	syncloudAccount, err := asserts.Decode([]byte(encodedSyncloudAccount))
	if err != nil {
		panic(fmt.Sprintf("cannot decode trusted assertion: %v", err))
	}
	syncloudRootAccountKey, err := asserts.Decode([]byte(encodedSyncloudRootAccountKey))
	if err != nil {
		panic(fmt.Sprintf("cannot decode trusted assertion: %v", err))
	}
	trustedAssertions = []asserts.Assertion{canonicalAccount, canonicalRootAccountKey, syncloudAccount, syncloudRootAccountKey}
}

// Trusted returns a copy of the current set of trusted assertions as used by Open.
func Trusted() []asserts.Assertion {
	trusted := []asserts.Assertion(nil)
	if !snapdenv.UseStagingStore() {
		trusted = append(trusted, trustedAssertions...)
	} else {
		if len(trustedStagingAssertions) == 0 {
			panic("cannot work with the staging store without a testing build with compiled-in staging keys")
		}
		trusted = append(trusted, trustedStagingAssertions...)
	}
	trusted = append(trusted, trustedExtraAssertions...)
	return trusted
}

// InjectTrusted injects further assertions into the trusted set for Open.
// Returns a restore function to reinstate the previous set. Useful
// for tests or called globally without worrying about restoring.
func InjectTrusted(extra []asserts.Assertion) (restore func()) {
	prev := trustedExtraAssertions
	trustedExtraAssertions = make([]asserts.Assertion, len(prev)+len(extra))
	copy(trustedExtraAssertions, prev)
	copy(trustedExtraAssertions[len(prev):], extra)
	return func() {
		trustedExtraAssertions = prev
	}
}
