/**
 * Copyright (c) 2025 worldiety GmbH
 *
 * This file is part of the NAGO Low-Code Platform.
 * Licensed under the terms specified in the LICENSE file.
 *
 * SPDX-License-Identifier: Custom-License
 */

export class BinaryWriter {
    private buffer: Uint8Array;
    private view: DataView;
    private offset: number = 0;
    private tmp: Uint8Array;

    constructor(size: number = 1024) {
        this.buffer = new Uint8Array(size);
        this.tmp = new Uint8Array(8);
        this.view = new DataView(this.buffer.buffer);
    }

    private ensureCapacity(additionalBytes: number) {
        if (this.offset + additionalBytes > this.buffer.length) {
            let nextSize = Math.max(additionalBytes + this.offset * 2, this.buffer.length * 2)
            const newBuffer = new Uint8Array(nextSize);
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

    writeFloat64(value: number): void {
        const buffer = new ArrayBuffer(8);
        const float64Bits = new DataView(buffer);
        float64Bits.setFloat64(0, value, true);
        this.tmp.set(new Uint8Array(buffer));
        this.write(this.tmp);
    }
}

export class BinaryReader {
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

    // Reads a specific number of bytes from the buffer
    readBytes(length: number): Uint8Array {
        if (this.offset + length > this.buffer.length) {
            throw new Error("Attempt to read beyond buffer length");
        }

        const bytes = this.buffer.slice(this.offset, this.offset + length);
        this.offset += length;
        return bytes;
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
        return {shape: header.shape, typeId};
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

    readVarint(): number {
        const uvalue = this.readUvarint();
        return (uvalue >>> 1) ^ -(uvalue & 1);
    }


    readFloat64(): number {
        let tmp = this.readBytes(8);
        const buffer = tmp.buffer;
        const view = new DataView(buffer);
        return view.getFloat64(0, true);
    }
}

// Types and Enums

export type Shape = number;
export type FieldId = number;
export type TypeId = number;

export enum Shapes {
    ENVELOPE = 0,
    UVARINT,
    VARINT,
    BYTESLICE,
    RECORD,
    F32,
    F64,
    ARRAY,
}

export interface FieldHeader {
    shape: Shape;
    fieldId: FieldId;
}

// Interface for writable objects
export interface Writeable {
    write(writer: BinaryWriter): void;

    writeTypeHeader(dst: BinaryWriter): void
}


// Interface for readable objects
export interface Readable {
    read(reader: BinaryReader): void;

    isZero(): boolean;
}


function writeString(writer: BinaryWriter, value: string): void {
    const data = new TextEncoder().encode(value); // Convert string to Uint8Array
    writer.writeUvarint(data.length); // Write the length of the string
    writer.write(data); // Write the string data
}

function readString(reader: BinaryReader): string {
    const strLen = reader.readUvarint(); // Read the length of the string
    const buf = reader.readBytes(strLen); // Read the string data
    return new TextDecoder().decode(buf); // Convert Uint8Array to string
}


function writeInt(writer: BinaryWriter, value: number): void {
    writer.writeUvarint(value);
}

function readInt(reader: BinaryReader): number {
    return reader.readUvarint();
}

function writeSint(writer: BinaryWriter, value: number): void {
    writer.writeVarint(value);
}

function readSint(reader: BinaryReader): number {
    return reader.readVarint();
}


function writeBool(writer: BinaryWriter, value: boolean): void {
    writer.writeUvarint(value ? 1 : 0);
}

function readBool(reader: BinaryReader): boolean {
    return reader.readUvarint() === 1;
}

function writeFloat(writer: BinaryWriter, value: number): void {
    writer.writeFloat64(value);
}

function readFloat(reader: BinaryReader): number {
    return reader.readFloat64();
}
