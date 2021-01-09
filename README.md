# BLEAS v2

"ah, bless" - when used patronisingly, meaning someone tried to do the right thing but didn't.

written in go.

## Bluetooth Low Energy Advertisement Sniffer

Based on the great work on the following projects:

- [sausheong](https://towardsdatascience.com/spelunking-bluetooth-le-with-go-c2cff65a7aca) spelunking for bluetooth (in go)
- [BLEAK](https://github.com/hbldh/bleak) Bluetooth Low Energy platform Agnostic Klient for Python
- [furiousMAC](https://github.com/furiousMAC/continuity) Apple Continuity Protocol Reverse Engineering and Dissector
- [popets 2020-0003](https://content.sciendo.com/view/journals/popets/2020/1/article-p26.xml?language=en) Discontinued Privacy: Personal Data Leaks in
Apple Bluetooth-Low-Energy Continuity
Protocols

## ElasticSearch

TODO: All data is saved into an ElasticSearch cluster, this enables better search and visualisation of data

## Privacy

It is possible to track users with this data as can be seen in the predecessor works, though no attempt has been to do this, and others using this code should not exercise this feature.

## Remove sudo requirement

`sudo setcap 'cap_net_raw,cap_net_admin+eip' scanner`

Obviously this is a security risk.
