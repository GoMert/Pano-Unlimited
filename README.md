# Pano

Windows için gelişmiş pano yöneticisi. Kopyaladığınız metin ve görseller kaybolmaz, şifreli olarak saklanır.

## Özellikler

- Metin ve görsel desteği
- AES-256 şifreleme (donanım tabanlı anahtar)
- Önemli öğeleri sabitleme
- Ctrl+Shift+V ile hızlı erişim
- System tray entegrasyonu
- Açık/koyu tema
- Windows başlangıcında otomatik çalışma

## Kurulum

### Hazır Exe

[Releases](../../releases) sayfasından son sürümü indirin.

### Kaynak Koddan Derleme

Go 1.21+ gereklidir.

```
go install github.com/tc-hib/go-winres@latest
go-winres make
go build -ldflags="-H windowsgui" -o Pano.exe .
```

## Kullanım

- Uygulama arka planda çalışır
- `Ctrl+Shift+V` ile pano penceresini açın/kapatın
- System tray ikonuna sağ tıklayarak menüye erişin
- Öğelere tıklayarak kopyalayın veya sabitleyin

## Lisans

MIT
