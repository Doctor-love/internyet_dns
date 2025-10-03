# internyet\_dns

## Introduction
This service enables party participants to register/configure arbitrary DNS
sub-domains under their own "namespace" (\*.\$ALIAS.p.nyet). If they are member
of a great house, it is also possible to manage sub-domains under the "group
namespace" (\*.\$GREAT\_HOUSE\_NAME.g.nyet)
  
Records can be defined by specifying an IPv4 address for A records or an IPv6
address for AAAA records. The special value "this" will use the client's
source address ("dynamic DNS"), but make sure to use the sub-domain
"v4.dns.c.nyet" or "v6.dns.c.nyet" for this to work properly.  
  
Users authentication using their assigned client certificate.  
  
Behind the scene, it just creates "hosts" files for dnsmasq to consume.
It relies on internyet\_proxy to authenticate users and should not be
exposed directly.


## Example usage
```
# Register A record for the sub-domain
# www.$ALIAS.p.nyet
# to 10.13.37.42

$ curl \
  --cacert /home/haxor/x509/ca.crt \
  --cert /home/haxor/x509/client.crt \
  --key /home/haxor/x509/client.key \
  --request POST \
  --header "X-SillyCSRF: false" \
  https://dns.c.nyet/api/v2/A/www/10.13.37.42

$ host www.darkdagger.p.nyet

www.darkdagger.p.nyet has address 10.13.37.42

# Register A record for the sub-domain
# deck.$ALIAS.p.nyet
# to the client's source IP ("dynamic DNS")

$ curl \
  --cacert /home/haxor/x509/ca.crt \
  --cert /home/haxor/x509/client.crt \
  --key /home/haxor/x509/client.key \
  --request POST \
  --header "X-SillyCSRF: false" \
  https://v4.dns.c.nyet/api/v2/A/deck/this

$ host deck.darkdagger.p.nyet

deck.darkdagger.p.nyet has address 10.13.37.105

# Register A record for the great house sub-domain
# ctf.legrup.g.nyet
# to 10.13.37.13

$ curl \
  --cacert /home/haxor/x509/ca.crt \
  --cert /home/haxor/x509/client.crt \
  --key /home/haxor/x509/client.key \
  --request POST \
  --header "X-SillyCSRF: false" \
  https://dns.c.nyet/api/v2/A/ctf/legrup/10.13.37.13

$ host ctf.legrup.g.nyet

ctf.legrup.g.nyet has address 10.13.37.13
```
