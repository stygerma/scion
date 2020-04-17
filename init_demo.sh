cp go/border/qos/testdata/DemoConfig.yaml gen/ISD1/ASff00_0_110/br1-ff00_0_110-1/qosConfig.yaml

cp go/border/qos/testdata/DemoConfigEmpty.yaml gen/ISD1/ASff00_0_111/br1-ff00_0_111-1/qosConfig.yaml

cp go/border/qos/testdata/DemoConfigEmpty.yaml gen/ISD1/ASff00_0_112/br1-ff00_0_112-1/qosConfig.yaml

./scion.sh stop && rm logs/* && make gazelle && ./scion.sh start && ./scion.sh status && sleep 5 && ./bin/scmp echo -local 1-ff00:0:110,[127.0.0.1] -remote 1-ff00:0:111,[0.0.0.0] -sciond 127.0.0.11:30255 && ./scion.sh stop