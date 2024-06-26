/*
 * Copyright 2018 Venafi, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *  http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package cloud

import (
	"net/http"
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Venafi/vcert/v5/pkg/certificate"
	"github.com/Venafi/vcert/v5/pkg/endpoint"
	"github.com/Venafi/vcert/v5/pkg/verror"
)

var (
	successGetUserAccount      = []byte("{\"user\": {\"username\": \"ben.skolmoski@venafi.com\",\"id\": \"aa4a4ee0-efaf-11e5-b223-d96cf8021ce5\",\"companyId\": \"a94d5140-efaf-11e5-b223-d96cf8021ce5\",\"userType\": \"EXTERNAL\",\"userAccountType\": \"API\",\"userStatus\": \"ACTIVE\",\"creationDate\": \"2016-03-21T21:55:45.998+0000\"},\"company\": {\"id\": \"a94d5140-efaf-11e5-b223-d96cf8021ce5\",\"companyType\": \"TPP_CUSTOMER\",\"active\": true,\"creationDate\": \"2016-03-21T21:55:44.326+0000\",\"domains\": [\"venafi.com\"]},\"apiKey\": {\"username\": \"ben.skolmoski@venafi.com\",\"apiType\": \"ALL\",\"apiVersion\": \"ALL\",\"apiKeyStatus\": \"ACTIVE\",\"creationDate\": \"2016-03-21T21:55:45.998+0000\"}}")
	errorGetUserAccount        = []byte("{\"errors\": [{\"code\": 10501,\"message\": \"Unable to find api key for key cec682ba-f409-40c0-9b00-aeb67876b7a1\",\"args\": [\"cec682ba-f409-40c0-9b00-aeb67876b7a1\"]}]}")
	successGetZoneByTag        = []byte("{\"id\": \"700e6820-0a60-11e7-a0e2-77cf2c42e000\",\"companyId\": \"700c4540-0a60-11e7-a0e2-77cf2c42e000\",\"tag\": \"Default\",\"zoneType\": \"OTHER\",\"certificatePolicyIds\": {\"CERTIFICATE_IDENTITY\": [\"700df2f0-0a60-11e7-a0e2-77cf2c42e000\"],\"CERTIFICATE_USE\": [\"700df2f1-0a60-11e7-a0e2-77cf2c42e000\"]},\"defaultCertificateIdentityPolicyId\": \"700df2f0-0a60-11e7-a0e2-77cf2c42e000\",  \"defaultCertificateUsePolicyId\": \"700df2f1-0a60-11e7-a0e2-77cf2c42e000\",\"systemGenerated\": true,\"creationDate\": \"2017-03-16T15:51:37.108+0000\"}")
	errorGetZoneByTag          = []byte("{\"errors\": [{\"code\": 10803,\"message\": \"Unable to find zone with tag Defaultwer for companyId 700c4540-0a60-11e7-a0e2-77cf2c42e000\",\"args\": [\"Defaultwer\",\"700c4540-0a60-11e7-a0e2-77cf2c42e000\"]}]}")
	successRequestCertificate  = []byte("{\"certificateRequests\": [{\"id\": \"04c051d0-f118-11e5-8b33-d96cf8021ce5\",\"zoneId\": \"a94d9f60-efaf-11e5-b223-d96cf8021ce5\",\"status\": \"ISSUED\",\"subjectDN\": \"cn=vcert.test.vfidev.com,ou=Automated Tests,o=Venafi, Inc.\",\"generatedKey\": false,\"defaultKeyPassword\": true,\"certificateInstanceIds\": [\"04bad390-f118-11e5-8b33-d96cf8021ce5\",\"04bcf670-f118-11e5-8b33-d96cf8021ce5\",\"04bf4060-f118-11e5-8b33-d96cf8021ce5\"],\"creationDate\": \"2016-03-23T16:55:16.589+0000\",\"pem\": \"-----BEGIN CERTIFICATE-----\\nMIIEkDCCA3igAwIBAgIBAjANBgkqhkiG9w0BAQsFADCBjTELMAkGA1UEBhMCQVUx\\nKDAmBgNVBAoMH1RoZSBMZWdpb24gb2YgdGhlIEJvdW5jeSBDYXN0bGUxIzAhBgNV\\nBAsMGkJvdW5jeSBQcmltYXJ5IENlcnRpZmljYXRlMS8wLQYJKoZIhvcNAQkBFiBm\\nZWVkYmFjay1jcnlwdG9AYm91bmN5Y2FzdGxlLm9yZzAeFw0xNTEyMjQxNjU1MTVa\\nFw0xNjA2MjExNjU1MTVaMIGSMQswCQYDVQQGEwJBVTEoMCYGA1UECgwfVGhlIExl\\nZ2lvbiBvZiB0aGUgQm91bmN5IENhc3RsZTEoMCYGA1UECwwfQm91bmN5IEludGVy\\nbWVkaWF0ZSBDZXJ0aWZpY2F0ZTEvMC0GCSqGSIb3DQEJARYgZmVlZGJhY2stY3J5\\ncHRvQGJvdW5jeWNhc3RsZS5vcmcwggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEK\\nAoIBAQDXmXlWdmym4BrYAGiwjf4lpZCuAa1trEoYZpoYbBrbSxTTc9YJPbZTdwf2\\nGfpu5Gtu6Aa4jBoLSB+hNqs9ViHMPvWelhfwR16EL4Zquz9O0RrigzDyEr0xTkhd\\no/1B8izZRzyEkfvHCi0j8fUtI/zS8XX5dXw+SJfHhdc8qTT7Hduczrio4QK2Y+bE\\nVSYUuie/RODVFRNNEUarBfnIChVyPM+ea+1AuGQaH1dw2Wjkkh90dFZYxNqg/DNF\\nMUsNLglLwIcKIjT5vQQg0ICcljXsrxYRCf2Cpw03GqVMARakSzbQTdrW7jKgue8T\\nqDD9BZaJhL4Vx3VwVGl5KYP2fsDjAgMBAAGjgfMwgfAwHQYDVR0OBBYEFMvG5bA8\\npPyoQq727jNHLkmbpecBMIG6BgNVHSMEgbIwga+AFMvG5bA8pPyoQq727jNHLkmb\\npecBoYGTpIGQMIGNMQswCQYDVQQGEwJBVTEoMCYGA1UECgwfVGhlIExlZ2lvbiBv\\nZiB0aGUgQm91bmN5IENhc3RsZTEjMCEGA1UECwwaQm91bmN5IFByaW1hcnkgQ2Vy\\ndGlmaWNhdGUxLzAtBgkqhkiG9w0BCQEWIGZlZWRiYWNrLWNyeXB0b0Bib3VuY3lj\\nYXN0bGUub3JnggEBMBIGA1UdEwEB/wQIMAYBAf8CAQAwDQYJKoZIhvcNAQELBQAD\\nggEBAJK16ApH7IU+nJ0gWNNncEWV8BBqTdsYPix1sKiMlNZzXG8l9M2DzVSbUBoF\\nZ63QDqp1VUlUX1N11b074tGr2JBmZNSRDaj61qRLKqWbcKlSeWAOwrzyeBUJWR5N\\nfMl/pE19uGcf4L0/SMPDboTiytTDGV/AszhAnsXVm/J4H27C4fPQVl3z0NY1VxnN\\nvbuD+qNRIYEbHpmpRzwpVDPIL3Qsp5AGq2Zeci9tr2F8aEl5EAxMcbT5FBZ6R9+B\\nhGAAId1NZgE3Xndt41KcgLPitNJ5ClSDecFU+gW0l3yv8/xBPBcAzoxUWW8Q32jJ\\ncneyJVzgKNLi7RBAMufyvql+6P8=\\n-----END CERTIFICATE-----\\n-----BEGIN CERTIFICATE-----\\nMIIDkDCCAngCAQEwDQYJKoZIhvcNAQELBQAwgY0xCzAJBgNVBAYTAkFVMSgwJgYD\\nVQQKDB9UaGUgTGVnaW9uIG9mIHRoZSBCb3VuY3kgQ2FzdGxlMSMwIQYDVQQLDBpC\\nb3VuY3kgUHJpbWFyeSBDZXJ0aWZpY2F0ZTEvMC0GCSqGSIb3DQEJARYgZmVlZGJh\\nY2stY3J5cHRvQGJvdW5jeWNhc3RsZS5vcmcwHhcNMTUxMjI0MTY1NTE1WhcNMTYw\\nNjIxMTY1NTE1WjCBjTELMAkGA1UEBhMCQVUxKDAmBgNVBAoMH1RoZSBMZWdpb24g\\nb2YgdGhlIEJvdW5jeSBDYXN0bGUxIzAhBgNVBAsMGkJvdW5jeSBQcmltYXJ5IENl\\ncnRpZmljYXRlMS8wLQYJKoZIhvcNAQkBFiBmZWVkYmFjay1jcnlwdG9AYm91bmN5\\nY2FzdGxlLm9yZzCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBANeZeVZ2\\nbKbgGtgAaLCN/iWlkK4BrW2sShhmmhhsGttLFNNz1gk9tlN3B/YZ+m7ka27oBriM\\nGgtIH6E2qz1WIcw+9Z6WF/BHXoQvhmq7P07RGuKDMPISvTFOSF2j/UHyLNlHPISR\\n+8cKLSPx9S0j/NLxdfl1fD5Il8eF1zypNPsd25zOuKjhArZj5sRVJhS6J79E4NUV\\nE00RRqsF+cgKFXI8z55r7UC4ZBofV3DZaOSSH3R0VljE2qD8M0UxSw0uCUvAhwoi\\nNPm9BCDQgJyWNeyvFhEJ/YKnDTcapUwBFqRLNtBN2tbuMqC57xOoMP0FlomEvhXH\\ndXBUaXkpg/Z+wOMCAwEAATANBgkqhkiG9w0BAQsFAAOCAQEAC39Y+ODklhaI4fLY\\noGUiCHgJqejn1wKeUhqua1NA/urChdYUenSUdwEcy2x9+gNNXAdg+XGJG3MNycaY\\nes24zvc2848QWwESOZ/hI/747JVNgnmfgF4DcIFlvDljN57B5nduYpE/1VwdIq5p\\n5/P+jEkY9rjWvEujm0t/hbaYdkkayXQJMaMQNV57AhnwaecPpqzBuhzZxL6vZjtl\\nGPBd6++GNe93+xRUyhWueVGUg5k+TupMZQXWrWfngh5lWfNqqUuvtgT+0U2Isgk+\\njYODaZA5rEOxqnyUrawuWaYUmf03ezPKssRb0fVE3GjM+dx2PhSuIPBnY3YkEaPo\\njwBAww==\\n-----END CERTIFICATE-----\\n-----BEGIN CERTIFICATE-----\\nMIIDxDCCAqygAwIBAgIFAOQPMskwDQYJKoZIhvcNAQELBQAwgZIxLzAtBgkqhkiG\\n9w0BCQEWIGZlZWRiYWNrLWNyeXB0b0Bib3VuY3ljYXN0bGUub3JnMSgwJgYDVQQL\\nDB9Cb3VuY3kgSW50ZXJtZWRpYXRlIENlcnRpZmljYXRlMSgwJgYDVQQKDB9UaGUg\\nTGVnaW9uIG9mIHRoZSBCb3VuY3kgQ2FzdGxlMQswCQYDVQQGEwJBVTAeFw0xNjAz\\nMjMxNjU1MTZaFw0xNjA2MjExNjU1MTZaMHAxDzANBgNVBAgTBk5ldmFkYTESMBAG\\nA1UEBxMJTGFzIFZlZ2FzMRUwEwYDVQQKEwxWZW5hZmksIEluYy4xGDAWBgNVBAsT\\nD0F1dG9tYXRlZCBUZXN0czEYMBYGA1UEAxMPdGVzdC52ZW5hZmkuY29tMIIBIjAN\\nBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAo9YR84fN1hiTaHjKH2ZKrTHYgJOS\\nSc9/mWw7eCbT7yYxs1TG97UnxngOUh7lIRRpUVpWZNXfWuam76GA6NfhQk75pQRi\\na24KV20z4v2ZDho785WxlvjXLTBu86XgQAZZ4DAjqit2LVcaJNgqxGK8R9fWmmUA\\noNyQRoPy2fRd5vPdZhbK37V+JsdF3KHSHUHkciNzKtJmVcCL1mO1FkiKkaHOo49x\\nzkjJM2ZbS4RjH5LAeJv/+gYCIhTkhjpeSM35dM4kp2xGQiIrAf8xrvfJpklAWtcS\\nZnfUTH6sRiSlOfZz09JvHUQhjZzwGvtcuetP7FASAeCgH5QWSZKNYVfsywIDAQAB\\no0IwQDAdBgNVHQ4EFgQUFTjuG78m12WBgi4Kzl/4QXu0x+IwHwYDVR0jBBgwFoAU\\ny8blsDyk/KhCrvbuM0cuSZul5wEwDQYJKoZIhvcNAQELBQADggEBAATw1x9+c3RY\\nFE1cxvGIr6hud324qVIW3mo2J/L6QoJns5ESxSoe+f6VjWxHsBGvSvhJxuQsLgUp\\nSnZvB86HxY/imUluBouP6ov+6yPet4E22N+AGPVYk0yddca9GguJQsIqZC1bEEHm\\nLLGYl1APJ8DuGQ0fos6q55seFVGBQAPlrod18wIutJYDKetnxYHv/4ZLELQKojF1\\n7s2i4LiVhTrIGjTH9sWiybup8vXY7iBcqPQOuop3Him7ODpPAm+RSx9//8wGyI9X\\n5Xchl8p8KStjxFkk6zGBTcqrY2HbRyZW8vu+Uqa+NITPAcnmYy9dVI00c2oodCtl\\n7fsDMneM/PI=\\n-----END CERTIFICATE-----\\n\"}]}")
	errorRequestCertificate    = []byte("{\"errors\": [{\"code\": 10702,\"message\": \"Unable to parse bytes of certificate signing request\",\"args\": [\"\"]}]}")
	successRetrieveCertificate = []byte("-----BEGIN CERTIFICATE-----\nMIIEkDCCA3igAwIBAgIBAjANBgkqhkiG9w0BAQsFADCBjTELMAkGA1UEBhMCQVUx\nKDAmBgNVBAoMH1RoZSBMZWdpb24gb2YgdGhlIEJvdW5jeSBDYXN0bGUxIzAhBgNV\nBAsMGkJvdW5jeSBQcmltYXJ5IENlcnRpZmljYXRlMS8wLQYJKoZIhvcNAQkBFiBm\nZWVkYmFjay1jcnlwdG9AYm91bmN5Y2FzdGxlLm9yZzAeFw0xNTEyMjQyMDQ4MzJa\nFw0xNjA2MjEyMDQ4MzJaMIGSMQswCQYDVQQGEwJBVTEoMCYGA1UECgwfVGhlIExl\nZ2lvbiBvZiB0aGUgQm91bmN5IENhc3RsZTEoMCYGA1UECwwfQm91bmN5IEludGVy\nbWVkaWF0ZSBDZXJ0aWZpY2F0ZTEvMC0GCSqGSIb3DQEJARYgZmVlZGJhY2stY3J5\ncHRvQGJvdW5jeWNhc3RsZS5vcmcwggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEK\nAoIBAQDXmXlWdmym4BrYAGiwjf4lpZCuAa1trEoYZpoYbBrbSxTTc9YJPbZTdwf2\nGfpu5Gtu6Aa4jBoLSB+hNqs9ViHMPvWelhfwR16EL4Zquz9O0RrigzDyEr0xTkhd\no/1B8izZRzyEkfvHCi0j8fUtI/zS8XX5dXw+SJfHhdc8qTT7Hduczrio4QK2Y+bE\nVSYUuie/RODVFRNNEUarBfnIChVyPM+ea+1AuGQaH1dw2Wjkkh90dFZYxNqg/DNF\nMUsNLglLwIcKIjT5vQQg0ICcljXsrxYRCf2Cpw03GqVMARakSzbQTdrW7jKgue8T\nqDD9BZaJhL4Vx3VwVGl5KYP2fsDjAgMBAAGjgfMwgfAwHQYDVR0OBBYEFMvG5bA8\npPyoQq727jNHLkmbpecBMIG6BgNVHSMEgbIwga+AFMvG5bA8pPyoQq727jNHLkmb\npecBoYGTpIGQMIGNMQswCQYDVQQGEwJBVTEoMCYGA1UECgwfVGhlIExlZ2lvbiBv\nZiB0aGUgQm91bmN5IENhc3RsZTEjMCEGA1UECwwaQm91bmN5IFByaW1hcnkgQ2Vy\ndGlmaWNhdGUxLzAtBgkqhkiG9w0BCQEWIGZlZWRiYWNrLWNyeXB0b0Bib3VuY3lj\nYXN0bGUub3JnggEBMBIGA1UdEwEB/wQIMAYBAf8CAQAwDQYJKoZIhvcNAQELBQAD\nggEBANAkqjVJZKQ8mxgbicKKHuVPoVohaRsIAW6LnDVVITYqtACpdbCRb0EruaCy\n6lH188gAMXVmEpCrjW4wplhikFYdGJVdz5q/sR/hlYKTSuLq33f+GlMWUy0p3MwX\nQQC73IEv0gwAp9PVf8bI7zICzy1qzmvRY4PZTWnzq8zupyz1Srfwuj3YiQA0orr7\nMGvISJjLShfW65HvEqvH65TGOMCQrLERELBHyyVRJNh/6snU6ZHiWKHsUmMgclA7\nYzVMYjAopRTBQp171d3m7RPkj7OnxNDAiIlgW8jwVhZ+RRlmGGL64pz4EBzXQ51f\nnPeVyjisi4ewoFl4kPdW1OVvLUk=\n-----END CERTIFICATE-----\n-----BEGIN CERTIFICATE-----\nMIIDkDCCAngCAQEwDQYJKoZIhvcNAQELBQAwgY0xCzAJBgNVBAYTAkFVMSgwJgYD\nVQQKDB9UaGUgTGVnaW9uIG9mIHRoZSBCb3VuY3kgQ2FzdGxlMSMwIQYDVQQLDBpC\nb3VuY3kgUHJpbWFyeSBDZXJ0aWZpY2F0ZTEvMC0GCSqGSIb3DQEJARYgZmVlZGJh\nY2stY3J5cHRvQGJvdW5jeWNhc3RsZS5vcmcwHhcNMTUxMjI0MjA0ODMyWhcNMTYw\nNjIxMjA0ODMyWjCBjTELMAkGA1UEBhMCQVUxKDAmBgNVBAoMH1RoZSBMZWdpb24g\nb2YgdGhlIEJvdW5jeSBDYXN0bGUxIzAhBgNVBAsMGkJvdW5jeSBQcmltYXJ5IENl\ncnRpZmljYXRlMS8wLQYJKoZIhvcNAQkBFiBmZWVkYmFjay1jcnlwdG9AYm91bmN5\nY2FzdGxlLm9yZzCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBANeZeVZ2\nbKbgGtgAaLCN/iWlkK4BrW2sShhmmhhsGttLFNNz1gk9tlN3B/YZ+m7ka27oBriM\nGgtIH6E2qz1WIcw+9Z6WF/BHXoQvhmq7P07RGuKDMPISvTFOSF2j/UHyLNlHPISR\n+8cKLSPx9S0j/NLxdfl1fD5Il8eF1zypNPsd25zOuKjhArZj5sRVJhS6J79E4NUV\nE00RRqsF+cgKFXI8z55r7UC4ZBofV3DZaOSSH3R0VljE2qD8M0UxSw0uCUvAhwoi\nNPm9BCDQgJyWNeyvFhEJ/YKnDTcapUwBFqRLNtBN2tbuMqC57xOoMP0FlomEvhXH\ndXBUaXkpg/Z+wOMCAwEAATANBgkqhkiG9w0BAQsFAAOCAQEA0qoqGT/iiNufZGyg\nusTCEQkLETXvzeQNnTP3ZOo33zuTrkNzLXB7wufwSUCpJ4A7iJudfejCNF0PW0Tp\nMxI7ImhJrTDkrJrTSS2nrXmqiy73aNAt8yCF8w9yGA0tBbmsenh1vPweZkYT9vt/\nqaDuOzEtxgAW0pLTwO1VO3V0FebMtXVZqWebJYCR4MTwV87p/dYcU12d5DBvV0FX\n57f+e3yqk8hqNC0m7yPwQmwvwu6qrjhQf9u92otUr0wnJ+TPAXg3gN0dJWvmT8F6\nGOSSVqoWgUz7ma3+tI+D5z1s6NYqNljAAUQsXgC8s/7uek7b2eeFsewq692ZBawO\nE5bWEQ==\n-----END CERTIFICATE-----\n-----BEGIN CERTIFICATE-----\nMIIDqDCCApCgAwIBAgIEWPizGTANBgkqhkiG9w0BAQsFADCBkjEvMC0GCSqGSIb3\nDQEJARYgZmVlZGJhY2stY3J5cHRvQGJvdW5jeWNhc3RsZS5vcmcxKDAmBgNVBAsM\nH0JvdW5jeSBJbnRlcm1lZGlhdGUgQ2VydGlmaWNhdGUxKDAmBgNVBAoMH1RoZSBM\nZWdpb24gb2YgdGhlIEJvdW5jeSBDYXN0bGUxCzAJBgNVBAYTAkFVMB4XDTE2MDMy\nMzIwNDgzM1oXDTE2MDYyMTIwNDgzM1owVTEVMBMGA1UEChMMVmVuYWZpLCBJbmMu\nMRgwFgYDVQQLEw9BdXRvbWF0ZWQgVGVzdHMxIjAgBgNVBAMTGWNlcnRhZmkudGVz\ndDMyLnZlbmFmaS5jb20wggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQDu\nFIDROdRJkw2gsE1YjyWy4s17PGNZpaXchpKq6eAmbK+O83ckE5tNT+pT+PbZA74D\nGQECiEIHIhazEm4Vq7H/fDVxvFAxpOU8SQ4tITJIvMhZaNIm1+luc8+YeSMEC8FG\n9BM/IloKPP5e7cjz6R0QRXh+SqMHpYzPcTakqnjXKoIYXnYafJylNVzH7hS7VNWC\nQv5lnKRy41gCxDwzF/rD101zHSs+5tmMNNr1UrBFs7ePpEnnTNIQP47zpSJewo5r\n2taRFp8ydyY3uTaf8pWkl+1iEyxvdNVDK6fyBGSdg1FQ3SifqIa8MrXlZeQv1FmN\nQpovuYA2sIbC3pD5JkhhAgMBAAGjQjBAMB0GA1UdDgQWBBSFtgPoILigM4FCd79u\nT7T1RhcZmTAfBgNVHSMEGDAWgBTLxuWwPKT8qEKu9u4zRy5Jm6XnATANBgkqhkiG\n9w0BAQsFAAOCAQEAmaJZKQ5VnDO4bwn5XofN+1OQ/5lwyXsUopRJ8ARy+p2gUqrX\nw4Xunznd1ZJsvSsUrDgMZvzqpVfR0NFCunHdhwl+peRRG+5U2JFleYq8UlK02Jil\nnvl3buyHD2Ejt0VbSPzWMSAiqTRM/qVRf9odOubXshxKz7S+JPjl2TmvtOEvNg9I\nJbI+qo8gm1LGQ6p3iVeG2UsMYzK4eF8eN/7YaVGf9LQ1SmdJKp6ZX1NA4OQA9D94\n1vh+hciO1LGIm1UgNAeas+/osN4ClAagA1HFmQ7aC6KY5PtfYjvfmvbCg+LZRomp\ncvLklAkilVvNX83ZnKL+trHpNCH2oe+3FnaV1w==\n-----END CERTIFICATE-----")
)

const (
	expectedURL = "https://api2.projectc.venafi.com/v1/"
)

func TestParseGetUserAccountData(t *testing.T) {
	reg, err := parseJSON[userDetails](successGetUserAccount, verror.ServerError)
	if err != nil {
		t.Fatalf("err is not nil, err: %s", err)
	}

	if reg.User.Username != "ben.skolmoski@venafi.com" {
		t.Fatalf("Registration username did not match expected value of ben.skolmoski@venafi.com, actual: %s", reg.User.Username)
	}
}

func TestParseBadAPIKeyError(t *testing.T) {
	_, err := parseUserDetailsResult(http.StatusOK, http.StatusPreconditionFailed, "Auth Error", errorGetUserAccount)
	if err == nil {
		t.Fatalf("err nil, expected error back")
	}
}

func TestParseZoneResponse(t *testing.T) {
	_, err := parseZoneConfigurationResult(http.StatusOK, "", successGetZoneByTag)
	if err != nil {
		t.Fatalf("err is not nil, err: %s", err)
	}

	_, err = parseZoneConfigurationResult(http.StatusNotFound, "Not Found", errorGetZoneByTag)
	if err == nil {
		t.Fatalf("err nil, expected error back")
	}
}

func Test_toPolicy(t *testing.T) {
	for _, test := range []struct {
		certTempl *certificateTemplate
		expPolicy endpoint.Policy
	}{
		{
			certTempl: &certificateTemplate{},
			expPolicy: endpoint.Policy{},
		},
		{
			certTempl: &certificateTemplate{
				SubjectCNRegexes: []string{"cn1", "cn2"},
				SubjectORegexes:  []string{"o1", "o2"},
				SubjectOURegexes: []string{"ou1", "ou2"},
				SubjectSTRegexes: []string{"st1", "st2"},
				SubjectLRegexes:  []string{"l1", "l2"},
				SubjectCValues:   []string{"c1", "c2"},

				SANRegexes:                          []string{"dns1", "dns2"},
				SanRfc822NameRegexes:                []string{"email1", "email2"},
				SanIpAddressRegexes:                 []string{"ip1", "ip2"},
				SanUniformResourceIdentifierRegexes: []string{"uri1", "uri2"},
			},
			expPolicy: endpoint.Policy{
				SubjectCNRegexes: []string{"^cn1$", "^cn2$"},
				SubjectORegexes:  []string{"^o1$", "^o2$"},
				SubjectOURegexes: []string{"^ou1$", "^ou2$"},
				SubjectSTRegexes: []string{"^st1$", "^st2$"},
				SubjectLRegexes:  []string{"^l1$", "^l2$"},
				SubjectCRegexes:  []string{"^c1$", "^c2$"},

				DnsSanRegExs:   []string{"^dns1$", "^dns2$"},
				EmailSanRegExs: []string{"^email1$", "^email2$"},
				IpSanRegExs:    []string{"^ip1$", "^ip2$"},
				UriSanRegExs:   []string{"^uri1$", "^uri2$"},
				UpnSanRegExs:   nil,
			},
		},
		{
			certTempl: &certificateTemplate{
				KeyReuse:   true,
				SANRegexes: []string{".*example.com"},
			},
			expPolicy: endpoint.Policy{
				DnsSanRegExs: []string{"^.*example.com$"},

				AllowKeyReuse:  true,
				AllowWildcards: true,
			},
		},
		{
			certTempl: &certificateTemplate{
				KeyTypes: []allowedKeyType{
					{
						KeyType:    "RSA",
						KeyLengths: []int{88888},
					},
				},
			},
			expPolicy: endpoint.Policy{
				AllowedKeyConfigurations: []endpoint.AllowedKeyConfiguration{
					{
						KeyType:  certificate.KeyTypeRSA,
						KeySizes: []int{88888},
					},
				},
			},
		},
		{
			certTempl: &certificateTemplate{
				KeyTypes: []allowedKeyType{
					{
						KeyType:   "EC",
						KeyCurves: []string{"P256", "P-384", "ED25519"},
					},
				},
			},
			expPolicy: endpoint.Policy{
				AllowedKeyConfigurations: []endpoint.AllowedKeyConfiguration{
					{
						KeyType: certificate.KeyTypeECDSA,
						KeyCurves: []certificate.EllipticCurve{
							certificate.EllipticCurveP256,
							certificate.EllipticCurveP384,
							certificate.EllipticCurveED25519,
						},
					},
				},
			},
		},
	} {
		policy := test.certTempl.toPolicy()
		require.Equal(t, test.expPolicy, policy)
	}
}

func TestUpdateRequest(t *testing.T) {
	req := certificate.Request{}
	req.Subject.CommonName = "vcert.test.vfidev.com"
	req.Subject.Organization = []string{"Venafi, Inc."}
	req.Subject.OrganizationalUnit = []string{"Automated Tests"}
	req.Subject.Locality = []string{"Las Vegas"}
	req.Subject.Province = []string{"Nevada"}
	req.Subject.Country = []string{"US"}

	zoneConfig := getZoneConfiguration(nil)

	zoneConfig.UpdateCertificateRequest(&req)
}

func TestGenerateRequest(t *testing.T) {

	keyTypeRSA := certificate.KeyTypeRSA
	keyTypeEC := certificate.KeyTypeECDSA
	keyTypeED25519 := certificate.KeyTypeED25519
	csrOriginServiceGenerated := certificate.ServiceGeneratedCSR

	cases := []struct {
		name          string
		keyType       *certificate.KeyType
		csrOrigin     *certificate.CSrOriginOption
		request       *certificate.Request
		expectedError string
	}{
		{
			"GenerateRequest-RSA-NotProvided",
			nil,
			nil,
			&certificate.Request{},
			"",
		},
		{
			"GenerateRequest-RSA",
			&keyTypeRSA,
			nil,
			&certificate.Request{},
			"",
		},
		{
			"GenerateRequest-EC",
			&keyTypeEC,
			nil,
			&certificate.Request{},
			"",
		},
		{
			"GenerateRequest-ED25519",
			&keyTypeED25519,
			nil,
			&certificate.Request{},
			"",
		},
		{
			"GenerateRequest-ED25519",
			&keyTypeED25519,
			&csrOriginServiceGenerated,
			&certificate.Request{},
			"ED25519 keys are not yet supported for Service Generated CSR",
		},
	}

	// filling every request
	for _, testCase := range cases {
		testCase.request.Subject.CommonName = "vcert.test.vfidev.com"
		testCase.request.Subject.Organization = []string{"Venafi, Inc."}
		testCase.request.Subject.OrganizationalUnit = []string{"Automated Tests"}
		testCase.request.Subject.Locality = []string{"Las Vegas"}
		testCase.request.Subject.Province = []string{"Nevada"}
		testCase.request.Subject.Country = []string{"US"}

		if testCase.keyType != nil {
			testCase.request.KeyType = *testCase.keyType
		}

		if testCase.csrOrigin != nil {
			testCase.request.CsrOrigin = *testCase.csrOrigin
		}
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {

			zoneConfig := getZoneConfiguration(nil)

			req := c.request
			zoneConfig.UpdateCertificateRequest(req)

			conn := Connector{}
			err := conn.GenerateRequest(zoneConfig, req)
			if err != nil {
				if c.expectedError != "" {
					regexErr := regexp.MustCompile(c.expectedError)
					if !regexErr.MatchString(err.Error()) {
						t.Fatalf("didn't get expected error, expected: %s, got: %s", c.expectedError, err.Error())
					}
				} else {
					t.Fatalf("err is not nil, err: %s", err)
				}
			} else {
				if c.expectedError != "" {
					t.Fatalf("got nil error, expected: %s", c.expectedError)
				}
			}
		})
	}
}

func TestParseCertificateRequestResponse(t *testing.T) {
	_, err := parseCertificateRequestResult(http.StatusCreated, "", successRequestCertificate)
	if err != nil {
		t.Fatalf("err is not nil, err: %s", err)
	}

	_, err = parseCertificateRequestResult(http.StatusBadRequest, "Bad Data", errorRequestCertificate)
	if err == nil {
		t.Fatalf("err nil, expected error back")
	}

	_, err = parseCertificateRequestResult(http.StatusGone, "Something unexpected", errorRequestCertificate)
	if err == nil {
		t.Fatalf("err nil, expected error back")
	}
}

func TestParseCertificateRetrieveResponse(t *testing.T) {
	_, err := newPEMCollectionFromResponse(successRetrieveCertificate, certificate.ChainOptionRootFirst)
	if err != nil {
		t.Fatalf("err is not nil, err: %s", err)
	}
}
