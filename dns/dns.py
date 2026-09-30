import ipaddress
from mitmproxy import dns, ctx

LOCAL = "[IP_REDACTED]"
# Critical domains to be spoofed
domains = {
    "axing.nintendowifi.net": LOCAL,
    "pubus-p.est.c.app.nintendowifi.net": LOCAL,
    "pubeu-p.est.c.app.nintendowifi.net": LOCAL,
    "pubus-p.est.c.app.pretendo.cc": LOCAL,
    "pubeu-p.est.c.app.pretendo.cc": LOCAL,
    "pubes-p.est.c.app.nintendowifi.net": LOCAL,
    "nppl.c.app.nintendowifi.net": LOCAL,
    "npul.c.app.nintendowifi.net": LOCAL,
    "npdl.c.app.nintendowifi.net": LOCAL,
    "npvk.app.pretendo.cc": LOCAL,
    "npvk.app.nintendo.net": LOCAL,
    "npdl.cdn.nintendowifi.net": LOCAL,
}

TTL = 60  

def lookup(name: str):
    name = name.lower().rstrip(".")
    for domain, ip in domains.items():
        if name == domain or name.endswith("." + domain):
            return ip
    return None


class PretendoDNS:
    def dns_request(self, flow: dns.DNSFlow) -> None:
        q = flow.request.question
        if q is None:
            return

        if q.type == dns.types.ANY:
            flow.response = flow.request.fail(dns.response_codes.REFUSED)
            return

        ip = lookup(q.name)
        if ip is None:
            return  # forward upstream

        if q.type == dns.types.A:
            rr = dns.ResourceRecord(
                name=q.name,
                type=dns.types.A,
                class_=dns.classes.IN,
                ttl=TTL,
                data=ipaddress.ip_address(ip).packed,
            )
            flow.response = flow.request.succeed([rr])
            ctx.log.info(f"Spoofed {q.name} -> {ip}")
        else:
            flow.response = flow.request.succeed([])


addons = [PretendoDNS()]
