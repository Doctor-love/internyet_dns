# internyet\_dns

## Introduction
This service enables party participants to register/configure arbitrary DNS
sub-domains under their own "namespace" (\*.\$ALIAS.e01.internyet.party).  
  
Records can be defined by specifying an IPv4 address for A records or an IPv6
address for AAAA records. The special value "this" will use the client's
source address ("dynamic DNS"), but make sure to use the sub-domain
"v4.dns.core.e01.internyet.party" or "v6.dns.core.e01.internyet.party" for
this to work properly.  
  
Users authentication using their assigned client certificate.  
  
Behind the scene, it just creates "hosts" files for dnsmasq to consume.


## Example usage
```
# Register A record for the sub-domain
# www.$ALIAS.e01.internyet.party
# to 10.13.37.42

$ curl \
  --cacert /home/haxor/x509/ca.crt \
  --cert /home/haxor/x509/client.crt \
  --key /home/haxor/x509/client.key \
  --request POST \
  --header "X-SillyCSRF: false" \
  https://dns.core.e01.internyet.party/api/v1/A/www/10.13.37.42

$ host www.darkdagger.e01.internyet.party

www.darkdagger.e01.internyet.party has address 10.13.37.42

# Register A record for the sub-domain
# deck.$ALIAS.e01.internyet.party
# to the client's source IP ("dynamic DNS")

$ curl \
  --cacert /home/haxor/x509/ca.crt \
  --cert /home/haxor/x509/client.crt \
  --key /home/haxor/x509/client.key \
  --request POST \
  --header "X-SillyCSRF: false" \
  https://v4.dns.core.e01.internyet.party/api/v1/A/deck/this

$ host deck.darkdagger.e01.internyet.party

deck.darkdagger.e01.internyet.party has address 10.13.37.105
```
