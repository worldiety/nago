# nprotoc

The NAGO protocol compiler is similar to protobuf, flatpak, captn proto, apache avro and many more.
However, it makes a few assumptions to make different tradeoffs.

It is mainly inspired by Avro, BACnet and MessagePack.
The following design goals are important:

* think of reducing all names in a JSON format to single digits. It is still valid and parseable and you only
  need the resolution table (the schema) to know which field is actually which index. But you can always understand
  the format, even numbers are missing or extra.
* provide a single self-contained marshal/unmarshal unit after compilation
* no dependencies beyond the target platforms standard library (or whatever is common)
* backwards and forwards parser compatibility
* (limited) schemaless inspection
* streamable
* KISS
* good compromise between speed and space
* nice for runtime parsers and static parsers
* allow zero alloc implementations
* clear and idiomatic code generation, no use of reflection or pointer escapades
* each field can be omitted, value is initialized using its natural zero value
* declarations are simply written in JSON, one file per type. The Filename is the type name (minus .json).
  No schema fuss.

The following limitations are important:

* there is no random access index, seeking requires more or less parsing everything. Actually, there is an
  envelope type, which may be injected to support parsers.
* there are no name spaces
* you cannot have more than 30 fields per record in the default 1 byte field header. But you can make them arbitrary larger by introducing more indirections
  and encapsulating and nest fields in more records. Index 0 is reserved to detect a type header and a future implementation can use index 31 to indicate an overflow node with a varuint following.
  The current implementation will reject to compile records with more than 30 fields today.
* there is no Nil or Null, if a field is not available, it must not be encoded. Details depend on the
  actual code generator, e.g. if a zero value is returned or an optional. Encoders are encouraged to omit zero values
  at all.
* even though, encoding most zero values is fine, an encoder should omit them, to minimize space and time consumption
* essentially, like JSON we have nested Objects with fields, arrays and primitives (which we call shapes)
* there are no human-readable type names or fields, instead there are memory shapes and schema defined types
  and fields.

## specification details

Each field has the segment number, followed immediately by the types data representation.

### field header

A field header for field ids upto 30 are always 1 byte large and encodes the shape and the field id.
The index 31 indicates a field id overflow and an uvarint is following.
Note, that a zero id indicates a type header instead.

```
                                             │                                                 
┌─────────────────┐┌────────────────────────┐ ┌─────────────────┐                              
│     3 bits      ││         5 bits         │││     uvarint     │                              
│  memory shape   ││     1-30: field id     │ │  field id > 30  │                              
│                 ││                        │││                 │                              
└─────────────────┘└────────────────────────┘ └─────────────────┘         overflow for idx 31, 
                                             └────────────────────────▶   not yet implemented  
                                                                                               
                                                                                                              
```

### type header

At least at the root level, there must be a custom header to declare a type.
Thus, we cannot use the field header.
Instead, it starts like a field header with the encoded shape but the field id part is zero followed by
an uvarint to refer to schema type.
This allows parsers to instantiate a concrete type.

```
┌─────────────────┐┌────────────────────────┐┌─────────────────┐
│     3 bits      ││         5 bits         ││     uvarint     │
│  memory shape   ││ always 0 to indicate a ││     type id     │
│                 ││          type          ││                 │
└─────────────────┘└────────────────────────┘└─────────────────┘
```



### memory shapes

The core idea is, that there is always a defined memory shape so that any decoder can understand and skip
unknown parts while parsing.
A shape value id is represented using the first 3 bits of an entry, which allows at most 8 shapes.
The available shapes are defined as follows:

| value | shape                                        |
|-------|----------------------------------------------|
| 0     | envelope (and extension types)               |
| 1     | uvarint (unsigned variable integer encoding) |
| 2     | varint (signed variable zigzag encoding      |
| 3     | byte-array                                   |
| 4     | record                                       |
| 5     | float32                                      |
| 6     | float64                                      |
| 7     | polymorphic-array                            |

### envelope

Envelopes helps to segment and skip forward chunks of data.
It trades buffer space at the encoder side to optimize decoders.
We do not use varuint encoding here, because a stream encoder will need to allocate a hole and jump back
to write the actual written size, after buffering.

```
                 │                                                    
┌──────────────┐    ┌──────────────┐ ┌──────────────┐ ┌──────────────┐
│    header    │ │  │  byte count  │ │extension type│ │ header[0]... │
│              │    │    uint32    │ │    1 byte    │ │              │
└──────────────┘ │  └──────────────┘ └──────────────┘ └──────────────┘
                                                                      
                 │                                                                                       
```

The extension type flag provides additional room for future type extensions using an indirection over
the envelope.
The following extensions are defined:

| value | extension shape                                          |
|-------|----------------------------------------------------------|
| 0     | envelope, headers and data will follow (byte count size) |

### record

```
                 │                                                                                            
┌──────────────┐    ┌──────────────┐ ┌────────────┐ ┌────────────┐┌────────────┐ ┌────────────┐ ┌────────────┐
│    header    │ │  │ field count  │ │ header[0]  │ │  data[0]   ││    ...     │ │header[n-1] │ │ data[n-1]  │
│              │    │uint8 (max 31)│ │            │ │            ││            │ │            │ │            │
└──────────────┘ │  └──────────────┘ └────────────┘ └────────────┘└────────────┘ └────────────┘ └────────────┘
                                                                                                              
                 │                                                                                                                                                                              
```

### polymorphic-array

Each array entry may have a different memory shape, to allow polymorphic value entries.
Thus, this is not as compact, as it could have been, especially for large arrays.
The overhead becomes significant, the smaller each array entry is.
But remember, that you may solve that even better at a higher level, e.g. use the []byte shape and
compress your data in a semantic way, e.g. use bitsets for flags or use RLE and delta encoding
for time series data with low variance.

```
                 │                                                                                            
┌──────────────┐    ┌──────────────┐ ┌────────────┐ ┌────────────┐┌────────────┐ ┌────────────┐ ┌────────────┐
│    header    │ │  │element count │ │ header[0]  │ │  data[0]   ││    ...     │ │header[n-1] │ │ data[n-1]  │
│              │    │   uvarint    │ │            │ │            ││            │ │            │ │            │
└──────────────┘ │  └──────────────┘ └────────────┘ └────────────┘└────────────┘ └────────────┘ └────────────┘
                                                                                                              
                 │                                                                                            
```

### byte-arrays resp. []byte

Byte arrays and strings share the same memory shape.
Strings must be encoded as a valid UTF8 sequence and parsers may reject or ignore invalid sequences, if they
got encoded or decoded that way.

```
                 │                                 
┌──────────────┐    ┌──────────────┐ ┌────────────┐
│    header    │ │  │  byte count  │ │  payload   │
│              │    │   uvarint    │ │   bytes    │
└──────────────┘ │  └──────────────┘ └────────────┘
                                                   
                 │                                 
```

### variable integer encodings

Just like the Go standard library does it which is probably equal to how protobuf encodes it.
There is unsigned variant and a signed variant using zigzag encoding.
The variable encoding is used for all kinds of integers, independent of their size (e.g. uint8 vs uint64).
Use this also for (an inefficient) representation of bool values (0=false and 1=true).

### float

The IEEE 754 binary representation is used for both float32 (4 byte) and float64 (8 byte) numbers.

## schema examples

A schema provides additional information to describe and document your binary format and the source code model to
be generated.
There are more semantic types, which will be mapped to the same memory shape.
For example strings and bytes slices or arrays are mapped to the byte-array shape.
A Map<string,string> is implicitly mapped to a record consisting of the according shapes (e.g. two strings without
any schema id).

### Enum

There is no enum memory shape, because arbitrary shape polymorphism is handled already at the header level.
Thus, these type information are only used to generate the according encoders and decoders.

```
{
  "doc": "Event is the union type of all allowed NAGO protocol events. Everything which goes through a NAGO channel must be an Event at the root level.",
  "type": "enum",
  "shape": "none",
  "id": "1",

  "variants": [
     "UpdateStateValueRequested",
     "FunctionCallRequested"
  ]
}
```

### Encoding efficiency examples

Even the smallest comparable json examples show huge wins of the binary format. 
Consider the following json

```json
{"t":1,"1":1,"2":2}
```

This will be binary marshalled with a 2 byte type header, 1 byte for the actual field count, 1 byte per field (=2) and 1 byte(=2) per payload
which totals at 7 bytes.