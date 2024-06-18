# Hacked CoreDNS for SELECT demo

This branch contains a fork of CoreDNS with a demo implementation of a potential DNS Resource Record type called **SELECT**.

## What is SELECT?

The **SELECT** type allows an ordinary DNS zone file to be used for DNS load balancing, traffic steering, or geotargeting.  Each **SELECT** record contains the name of a "selection policy" that is used to choose the answer, a "base name" on which all the possible responses hang with separate labels, and a list of the participating RR types.  It looks like this:

```DNS Zone
$ORIGIN example.com.

www            IN 300 AAAA 2001:db8::1
www            IN 300 SELECT "tiger-1" choices AAAA
eenie.choices  IN 300 AAAA 2001:db8::2
meenie.choices IN 300 AAAA 2001:db8::3
minie.choices  IN 300 AAAA 2001:db8::4
moe.choices    IN 300 AAAA 2001:db8::5
```

This zone file indicates that AAAA queries to "www.example.com" should be answered using the "tiger-1" selection policy to choose "eenie", "meenie", "minie", "moe", or a default option.  This zone file can be transferred from primary to secondary nameservers via standard AXFR, but the primary and secondary nameservers must agree on the meaning of "tiger-1" out of band.

In this implementation, selection policies correspond to [CoreDNS Metadata](https://coredns.io/plugins/metadata/) keys, so any CoreDNS plugin that adds metadata tags (e.g. [geoip](https://coredns.io/plugins/geoip/)) can be used as a selection policy.

## How to run the demo

```
git clone --branch bemasc-select --depth 1 --recurse-submodules https://github.com/bemasc/coredns.git
cd coredns
make
./coredns -conf plugin/file/select_example.corefile
```

# Things to try

```
$ dig +short @localhost -p 8053 TXT example.com
"udp (default)"
$ dig +short @localhost -p 8053 +tcp TXT example.com
"tcp"
$ dig +short @localhost -p 8053 +tcp TXT protocols.example.com
"tcp"
```

```
$ dig @localhost -p 8053 +short TXT cointoss.example.com
"head"
$ dig @localhost -p 8053 +short TXT cointoss.example.com
"tail"
```

```
$ dig +short @localhost -p 8053 +tcp TXT where.example.com
"Unknown"
% dig +short +subnet=123.45.67.0/24 @localhost -p 8053 TXT where.example.com
"Asia"
```

# Show me the code

* All diffs in this branch: https://github.com/bemasc/coredns/compare/master...bemasc-select
  - [Demo zone file](plugin/file/select_example.zone)
  - [Lookup engine](plugin/file/lookup.go)
  - [Selector scaffolding](plugin/file/select.go)
* SELECT RR type parser/serializer: https://github.com/bemasc/miekg-dns/compare/master...bemasc-select
