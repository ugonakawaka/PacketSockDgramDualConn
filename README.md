# PacketSockDgramDualConn

## 目的    

linux上でudpのconnを提供します。 
udpを通常扱う場合は、ipヘッダーは気にしないと思いますが、このconnではipヘッダーを取り扱います。  
実現方法は、AF_PACKET、SOCK_DGRAM、ETH_P_ALL、を指定することで実現しています。
ネットワークインターフェイスを指定せず、ipv4、ipv6の両方からのudpをひとつのconnで受け取れるように実装しています。  

## ちなみに
AF_PACKET、SOCK_DGRAM、ETH_P_ALLの指定だけでひとつのソケットで、対応できることはわかったのですが、udpをふたつ受けとってしまうため、このような実装にしました。  
プロジェクト名のDualConnは、仕方なく、ふたつのconnを使ってますよ、というニュアンスのつもりです。  
正確にはconnではないですが...  
ソースコードは単純にソケットつかってrecvしているだけです。  


## 疑問  
- golangの標準のnetで実現できるのでは？  
はい、その疑念はあります。RawConn使えば、よいのかなといろいろ試したのですが、うまくできませんでした。  
また、

- 性能でるの？  
わかりません。他の実装と比べきれていないので、正直わかりません、中の実装では、ひろえたデータグラムから自分宛てのポートかどうかでの判定を行っているため、そういった処理が、性能に影響するのだろうかですとか、他の要因があって実はよくないという疑念はあります。