# samsung-scm-dumper

Minimal Cobra-based CLI that reads a Samsung `.scm` archive and dumps satellite channels in the TV's current order.

## Build

```bash
# Go 1.22+
go mod tidy
go build -o channelsort
```

## Usage

```bash
./channelsort dump --scm /path/to/channel_list_XXXXX.scm
```

Output columns:

- **LCN**: Logical channel number from the record
- **SLOT**: Physical record index (0-based) inside `map-SateD`
- **NAME**: Service name (UTF-16BE decoded)

## Notes

- Parser is **non-destructive** and keeps raw bytes for future reordering tools.
- For your model, `map-SateD` uses **168-byte** fixed records with name at offset **0x24**.
- When you’re ready to implement reordering, prefer **swapping whole 168-byte records** to avoid touching unknown flags.
