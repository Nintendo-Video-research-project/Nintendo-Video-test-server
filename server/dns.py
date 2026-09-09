from ipaddress import ip_address
from mitmproxy import dns
import logging

# Critical domains to be spoofed
domains = {
    "pubus-p.est.c.app.nintendowifi.net",
    "nppl.c.app.nintendowifi.net",
    "npul.c.app.nintendowifi.net",
    "npvk.app.nintendo.net",
    "logus-p.est.c.app.nintendowifi.net",
}
#Your dns server ip
dns_server = ""  

def dns_request(flow: dns.DNSFlow) -> None:
    q = flow.request.question
    if q.name in domains and q.type == dns.types.A:
        logging.info(f"Spoofing DNS record for {q.name}")
        flow.response = flow.request.succeed([
            dns.ResourceRecord.A(
                name=q.name,
                ip=ip_address(dns_server),
            )
        ])
