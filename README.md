# route53dydndns

Yet another little tool to update a DNS record via Route53 for dynamic DNS purpose. You can
install it via

    go get github.com/tronical/route53dyndns

Call it with the host name as first parameter and the domain name as second, for example:

    route53dyndns mymachine mydomain.com

It uses check http://checkip.amazonaws.com/ to determine the IP address. Then it issues
an upsert request to route53 to create/update mymachine.mydomain.com with an DNS A record.

The calls to AWS requires credentials. The best way to configure credentials is to use the
`~/.aws/credentials` file, which might look like:

```
[default]
aws_access_key_id = AKID1234567890
aws_secret_access_key = MY-SECRET-KEY
```

Alternatively, you can set the following environment variables:

```
AWS_ACCESS_KEY_ID=AKID1234567890
AWS_SECRET_ACCESS_KEY=MY-SECRET-KEY
```

