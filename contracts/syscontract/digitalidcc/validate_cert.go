/*
 *
 *    This file is part of go-palletone.
 *    go-palletone is free software: you can redistribute it and/or modify
 *    it under the terms of the GNU General Public License as published by
 *    the Free Software Foundation, either version 3 of the License, or
 *    (at your option) any later version.
 *    go-palletone is distributed in the hope that it will be useful,
 *    but WITHOUT ANY WARRANTY; without even the implied warranty of
 *    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *    GNU General Public License for more details.
 *    You should have received a copy of the GNU General Public License
 *    along with go-palletone.  If not, see <http://www.gnu.org/licenses/>.
 * /
 *
 *  * @author PalletOne core developer <dev@pallet.one>
 *  * @date 2018-2019
 *
 */

package digitalidcc

import (
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"github.com/palletone/go-palletone/contracts/shim"
	dagConstants "github.com/palletone/go-palletone/dag/constants"
	"time"
)

// This is the basic validation
func ValidateCert(issuer string, cert *x509.Certificate, stub shim.ChaincodeStubInterface) error {
	if err := checkExists(cert, stub); err != nil {
		return err
	}
	if err := validateIssuer(issuer, cert, stub); err != nil {
		return err
	}
	// validate

	return nil
}

func ValidateCRLIssuer(issuer string, crl *pkix.CertificateList, stub shim.ChaincodeStubInterface) (certHolder []*CertHolderInfo, err error) {
	// check issuer identity
	certsInfo, err := getIssuerCertsInfo(issuer, stub)
	if err != nil {
		return nil, err
	}
	certHolder = []*CertHolderInfo{}
	for _, revokeCert := range crl.TBSCertList.RevokedCertificates {
		var i int = 0
		for j, holder := range certsInfo {
			i = j
			if revokeCert.SerialNumber.String() == holder.CertID {
				certHolder = append(certHolder, holder)
				break
			}
		}
		if i > len(certsInfo) {
			return nil, fmt.Errorf("Issuer(%s) can not revoke cert(%s): has no authority", issuer, revokeCert.SerialNumber.String())
		}
	}
	if len(certHolder) != len(crl.TBSCertList.RevokedCertificates) {
		return nil, fmt.Errorf("DigitalIdentityChainCode addCRLCert validate error: cert lenth is invalid")
	}
	return certsInfo, nil
}

func checkExists(cert *x509.Certificate, stub shim.ChaincodeStubInterface) error {
	// check root ca
	val, err := stub.GetSystemConfig("RootCABytes")
	if err != nil {
		return err
	}
	bytes, err := loadCertBytes([]byte(val))
	if err != nil {
		return err
	}
	rootCert, err := x509.ParseCertificate(bytes)
	if err != nil {
		return err
	}
	if rootCert.SerialNumber.String() == cert.SerialNumber.String() {
		return fmt.Errorf("Can not add root ca.")
	}
	// check other certificates
	key := dagConstants.CERT_BYTES_SYMBOL + cert.SerialNumber.String()
	data, err := stub.GetState(key)
	if err != nil {
		return err
	}
	if len(data) > 0 {
		return fmt.Errorf("Cert(%s) is existing.", cert.SerialNumber.String())
	}
	return nil
}

func validateIssuer(issuer string, cert *x509.Certificate, stub shim.ChaincodeStubInterface) error {
	// check with root ca holder
	rootCAHolder, err := stub.GetSystemConfig("RootCAHolder")
	if err != nil {
		return err
	}
	// check in intermediate certificate
	rootCert, err := GetRootCert(stub)
	if err != nil {
		return err
	}
	if issuer == rootCAHolder {
		if cert.Issuer.String() != rootCert.Subject.String() {
			return fmt.Errorf("cert issuer is invalid")
		}
	} else {
		// query certid
		certid, err := GetCertIDBySubject(cert.Issuer.String(), stub)
		if err != nil {
			return err
		}
		if certid == "" {
			return fmt.Errorf("Has no validate intermidate certificate")
		}
		// query server list
		revocationTime, err := GetCertRevocationTime(issuer, certid, stub)
		if err != nil {
			return err
		}
		if revocationTime.IsZero() || revocationTime.String() < time.Now().String() {
			return fmt.Errorf("Has no validate intermidate certificate")
		}
	}
	return nil
}

// This is the certificate chain validation
// To validate certificate chain signature
func ValidateCertChain(cert *x509.Certificate, stub shim.ChaincodeStubInterface) error {
	// query root ca cert bytes
	rootCABytes, err := stub.GetSystemConfig("RootCABytes")
	if err != nil {
		return err
	}
	val, err := loadCertBytes([]byte(rootCABytes))
	if err != nil {
		return err
	}
	rootCert, err := x509.ParseCertificate(val)
	// query intermidate cert bytes
	chancerts := []*x509.Certificate{}
	if cert.Issuer.String() != rootCert.Subject.String() {
		chancerts, err = GetIntermidateCertChains(cert, rootCert.Subject.String(), stub)
		if err != nil {
			return err
		}
	}
	// package x509.VerifyOptions, Intermediates and Roots field
	roots := x509.NewCertPool()
	roots.AddCert(rootCert)

	intermediates := x509.NewCertPool()
	for _, newCert := range chancerts {
		intermediates.AddCert(newCert)
	}
	opts := x509.VerifyOptions{
		Roots:         roots,
		Intermediates: intermediates,
	}
	// user x509.Verify to verify cert chain
	if _, err := cert.Verify(opts); err != nil {
		return err
	}

	return nil
}

// This is the certificate chain organization
