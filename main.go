/*
 Copyright (C) 2016 Simon Hausman <hausmann@gmail.com>

 Redistribution and use in source and binary forms, with or without
 modification, are permitted provided that the following conditions are
 met:
   * Redistributions of source code must retain the above copyright
     notice, this list of conditions and the following disclaimer.
   * Redistributions in binary form must reproduce the above copyright
     notice, this list of conditions and the following disclaimer in
     the documentation and/or other materials provided with the
     distribution.
   * Neither the name of The Qt Company Ltd nor the names of its
     contributors may be used to endorse or promote products derived
     from this software without specific prior written permission.


 THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
 "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
 LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
 A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
 OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
 SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
 LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
 DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
 THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
 (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
 OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
*/
package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/aws/aws-sdk-go/service/route53/route53iface"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

func getOwnIp() (string, error) {
	resp, err := http.Get("http://checkip.amazonaws.com/")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(buf)), nil
}

func findHostedZone(service route53iface.Route53API, domain string) (*string, error) {
	input := &route53.ListHostedZonesByNameInput{
		DNSName: aws.String(domain),
	}
	response, err := service.ListHostedZonesByName(input)
	if err != nil {
		return nil, nil
	}
	if len(response.HostedZones) != 1 {
		return nil, fmt.Errorf("Unexpected number of hosted zones found: %v expected 1", len(response.HostedZones))
	}
	return response.HostedZones[0].Id, nil
}

func updateRecordSet(service route53iface.Route53API, zoneId *string, name string, newIP string) error {
	change := &route53.ChangeResourceRecordSetsInput{
		HostedZoneId: zoneId,
		ChangeBatch: &route53.ChangeBatch{
			Changes: []*route53.Change{{
				Action: aws.String(route53.ChangeActionUpsert),
				ResourceRecordSet: &route53.ResourceRecordSet{
					Name: aws.String(name),
					Type: aws.String("A"),
					TTL:  aws.Int64(300),
					ResourceRecords: []*route53.ResourceRecord{{
						Value: aws.String(newIP),
					}},
				},
			}},
			Comment: aws.String("Update"),
		},
	}

	_, err := service.ChangeResourceRecordSets(change)
	if err != nil {
		return err
	}
	return nil
}

func appMain() error {
	if len(os.Args) != 3 {
		return fmt.Errorf("Usage: %s host-name domain\n", os.Args[0])
	}

	ip, err := getOwnIp()
	if err != nil {
		return err
	}

	svc := route53.New(session.New())
	name := os.Args[1]
	domain := os.Args[2]

	fqdn := name + "." + domain + "."

	zoneId, err := findHostedZone(svc, domain)
	if err != nil {
		return err
	}

	if err := updateRecordSet(svc, zoneId, fqdn, ip); err != nil {
		return err
	}

	return nil
}

func main() {
	if err := appMain(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
