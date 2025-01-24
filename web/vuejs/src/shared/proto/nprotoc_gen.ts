// Code generated by NAGO nprotoc DO NOT EDIT.


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
// NagoEvent is the union type of all allowed NAGO protocol events. Everything which goes through a NAGO channel must be an Event at the root level.
interface NagoEvent {
	// a marker method to indicate the enum / union type membership
	isNagoEvent(): void;
}

// Ptr represents an allocated instance within the backend which is unique in the associated scope.
class Ptr {
 
  private value: number; // Using number to handle uint64 (precision limits apply)

  constructor(value: number = 0) {
    this.value = value;
  }

  isZero(): boolean {
    return this.value === 0;
  }

  reset(): void {
    this.value = 0;
  }

  write(writer: BinaryWriter): void {
    writer.writeUvarint(this.value);
  }

  read(reader: BinaryReader): void {
    this.value = reader.readUvarint();
  }

}

// companion enum containing all defined constants for Ptr
enum PtrValues {
	// Null represents the zero value and a nil or null pointer address.
	Null = 0,
}



// UpdateStateValueRequested is raised from the frontend to update a state value hold by the backend. It can also immediately invoke a function callback in the same cycle.
class UpdateStateValueRequested implements NagoEvent {
	// The StatePointer must not be zero.
	statePointer: Ptr;

	// A FunctionPointer is invoked, if not zero.
	functionPointer: Ptr;

	constructor(statePointer: Ptr = new Ptr(), functionPointer: Ptr = new Ptr(), ) {
		this.statePointer = statePointer;
		this.functionPointer = functionPointer;
	}

	read(reader: BinaryReader): void {
		this.statePointer.reset();
		this.functionPointer.reset();
		const fieldCount = reader.readByte();
		for (let i = 0; i < fieldCount; i++) {
			const fieldHeader = reader.readFieldHeader();
			switch (fieldHeader.fieldId) {
				case 1:
					this.statePointer.read(reader);
					break
				case 2:
					this.functionPointer.read(reader);
					break
				default:
					throw new Error(`Unknown field ID: ${fieldHeader.fieldId}`);
			}
		}
	}

	write(writer: BinaryWriter): void {
		const fields = [false,!this.statePointer.isZero(),!this.functionPointer.isZero(),];
		let fieldCount = fields.reduce((count, present) => count + (present ? 1 : 0), 0);
		writer.writeByte(fieldCount);
		if (fields[1]) {
			writer.writeFieldHeader(Shapes.UVARINT, 1);
			this.statePointer.write(writer);
		}
		if (fields[2]) {
			writer.writeFieldHeader(Shapes.UVARINT, 2);
			this.functionPointer.write(writer);
		}
	}

	isNagoEvent(): void{}
}


// FunctionCallRequested tells the backend that the given pointer in the associated scope shall be invoked for a side effect.
class FunctionCallRequested implements NagoEvent {
	// Ptr denotes the remote pointer of the function.
	ptr: Ptr;

	constructor(ptr: Ptr = new Ptr(), ) {
		this.ptr = ptr;
	}

	read(reader: BinaryReader): void {
		this.ptr.reset();
		const fieldCount = reader.readByte();
		for (let i = 0; i < fieldCount; i++) {
			const fieldHeader = reader.readFieldHeader();
			switch (fieldHeader.fieldId) {
				case 1:
					this.ptr.read(reader);
					break
				default:
					throw new Error(`Unknown field ID: ${fieldHeader.fieldId}`);
			}
		}
	}

	write(writer: BinaryWriter): void {
		const fields = [false,!this.ptr.isZero(),];
		let fieldCount = fields.reduce((count, present) => count + (present ? 1 : 0), 0);
		writer.writeByte(fieldCount);
		if (fields[1]) {
			writer.writeFieldHeader(Shapes.UVARINT, 1);
			this.ptr.write(writer);
		}
	}

	isNagoEvent(): void{}
}


// Alignment is specified as follows:
// 
// 	┌─TopLeading───────────Top─────────TopTrailing─┐
// 	│                                              │
// 	│                                              │
// 	│                                              │
// 	│                                              │
// 	│                                              │
// 	│                                              │
// 	│                                              │
// 	│ Leading            Center            Trailing│
// 	│                                              │
// 	│                                              │
// 	│                                              │
// 	│                                              │
// 	│                                              │
// 	│                                              │
// 	│                                              │
// 	└BottomLeading───────Bottom──────BottomTrailing┘
// 
// An empty Alignment must be interpreted as Center (="c").
class Alignment {
 
  private value: number; // Using number to handle uint64 (precision limits apply)

  constructor(value: number = 0) {
    this.value = value;
  }

  isZero(): boolean {
    return this.value === 0;
  }

  reset(): void {
    this.value = 0;
  }

  write(writer: BinaryWriter): void {
    writer.writeUvarint(this.value);
  }

  read(reader: BinaryReader): void {
    this.value = reader.readUvarint();
  }

}

// companion enum containing all defined constants for Alignment
enum AlignmentValues {
		Center = 0,
		Top = 1,
		Bottom = 2,
		Leading = 3,
		Trailing = 4,
		TopLeading = 5,
		TopTrailing = 6,
		BottomLeading = 7,
		BottomTrailing = 8,
}


// Function to marshal a Writeable object into a BinaryWriter
function marshal(dst: BinaryWriter, src: Writeable): void {
	if (src instanceof Ptr) {
		dst.writeTypeHeader(Shapes.UVARINT, 2);
		src.write(dst);
		return
	}
	if (src instanceof UpdateStateValueRequested) {
		dst.writeTypeHeader(Shapes.RECORD, 3);
		src.write(dst);
		return
	}
	if (src instanceof FunctionCallRequested) {
		dst.writeTypeHeader(Shapes.RECORD, 4);
		src.write(dst);
		return
	}
	if (src instanceof Alignment) {
		dst.writeTypeHeader(Shapes.UVARINT, 5);
		src.write(dst);
		return
	}
}

// Function to unmarshal data from a BinaryReader into a Readable object
function unmarshal(src: BinaryReader): Readable {
	const { typeId } = src.readTypeHeader();
	switch (typeId) {
		case 2: {
			const v = new Ptr();
			v.read(src);
			return v;
		}
		case 3: {
			const v = new UpdateStateValueRequested();
			v.read(src);
			return v;
		}
		case 4: {
			const v = new FunctionCallRequested();
			v.read(src);
			return v;
		}
		case 5: {
			const v = new Alignment();
			v.read(src);
			return v;
		}
	}
	throw new Error(`Unknown type ID: ${typeId}`);
}

