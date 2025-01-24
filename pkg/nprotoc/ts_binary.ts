class BinaryWriter {
    private buffer: Uint8Array;
    private view: DataView;
    private offset: number = 0;

    constructor(size: number = 1024) {
        this.buffer = new Uint8Array(size);
        this.view = new DataView(this.buffer.buffer);
    }

    private ensureCapacity(additionalBytes: number) {
        if (this.offset + additionalBytes > this.buffer.length) {
            const newBuffer = new Uint8Array(this.buffer.length * 2);
            newBuffer.set(this.buffer);
            this.buffer = newBuffer;
            this.view = new DataView(this.buffer.buffer);
        }
    }

    write(p: Uint8Array) {
        this.ensureCapacity(p.length);
        this.buffer.set(p, this.offset);
        this.offset += p.length;
    }

    writeBool(b: boolean) {
        this.ensureCapacity(1);
        this.buffer[this.offset++] = b ? 1 : 0;
    }

    writeVarint(value: number) {
        this.ensureCapacity(10); // Maximum varint size is 10 bytes
        let v = value;
        while (v > 127) {
            this.buffer[this.offset++] = (v & 0x7f) | 0x80;
            v >>>= 7;
        }
        this.buffer[this.offset++] = v;
    }

    writeUvarint(value: number) {
        this.writeVarint(value >>> 0);
    }

    writeByte(b: number) {
        this.ensureCapacity(1);
        this.buffer[this.offset++] = b;
    }

    writeFieldHeader(shape: Shape, id: FieldId) {
        this.writeByte((shape << 5) | (id & 0b00011111));
    }

    writeTypeHeader(shape: Shape, id: TypeId) {
        this.writeFieldHeader(shape, 0);
        this.writeUvarint(id);
    }

    writeSlice(data: Uint8Array) {
        this.writeUvarint(data.length);
        this.write(data);
    }

    getBuffer(): Uint8Array {
        return this.buffer.slice(0, this.offset);
    }
}

class BinaryReader {
    private buffer: Uint8Array;
    private view: DataView;
    private offset: number = 0;

    constructor(buffer: Uint8Array) {
        this.buffer = buffer;
        this.view = new DataView(buffer.buffer);
    }

    readByte(): number {
        if (this.offset >= this.buffer.length) throw new Error("Out of bounds");
        return this.buffer[this.offset++];
    }

    readFieldHeader(): FieldHeader {
        const value = this.readByte();
        return {
            shape: (value >> 5) & 0b00000111,
            fieldId: value & 0b00011111,
        };
    }

    readTypeHeader(): { shape: Shape; typeId: TypeId } {
        const header = this.readFieldHeader();
        if (header.fieldId !== 0) {
            throw new Error("Expected type header, got field header");
        }
        const typeId = this.readUvarint();
        return { shape: header.shape, typeId };
    }

    readUvarint(): number {
        let result = 0;
        let shift = 0;
        while (true) {
            const byte = this.readByte();
            result |= (byte & 0x7f) << shift;
            if ((byte & 0x80) === 0) break;
            shift += 7;
            if (shift > 35) throw new Error("Varint too long");
        }
        return result;
    }
}

// Types and Enums

type Shape = number;
type FieldId = number;
type TypeId = number;

enum Shapes {
    ENVELOPE = 0,
    UVARINT,
    VARINT,
    BYTESLICE,
    RECORD,
    F32,
    F64,
    ARRAY,
}

interface FieldHeader {
    shape: Shape;
    fieldId: FieldId;
}

// Interface for writable objects
interface Writeable {
    write(writer: BinaryWriter): void;
}


// Interface for readable objects
interface Readable {
    read(reader: BinaryReader): void;
}