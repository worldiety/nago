// The runtime of the gostrict subset. Every name starts with $, which a Go identifier can not contain, so the
// translated code never collides with it. The translated code refers to the globals of JavaScript only through
// this runtime, so a Go variable may be called Math or Map.
//
// The checker proves, that a program can not fail at runtime: there is no nil, no index out of range and no
// division by zero. So the runtime contains no checks, only the semantics of Go, which JavaScript lacks.

const $M = Math;
const $B = BigInt;
const $N = Number;

// int64 and uint64 are BigInts, which each operation wraps around like Go

function $i64(x) {
	return $B.asIntN(64, x);
}

function $u64(x) {
	return $B.asUintN(64, x);
}

// $uf converts a uint64 into the nearest float64. Beyond the range of int64 it keeps the lost bit for the rounding,
// because goja converts a BigInt through int64.
function $uf(x) {
	return x < 9223372036854775808n ? $N(x) : $N((x >> 1n) | (x & 1n)) * 2;
}

function $imin(...a) {
	return a.reduce((x, y) => (y < x ? y : x));
}

function $imax(...a) {
	return a.reduce((x, y) => (y > x ? y : x));
}

// A variable shift count may be any unsigned value: from the width on, Go shifts all bits out.

function $shl32(x, n) {
	return n < 32 ? x << n : 0;
}

function $sar32(x, n) {
	return x >> (n < 32 ? n : 31);
}

function $shlu32(x, n) {
	return n < 32 ? (x << n) >>> 0 : 0;
}

function $shru32(x, n) {
	return n < 32 ? x >>> n : 0;
}

function $shl64(x, n) {
	return n < 64 ? $i64(x << $B(n)) : 0n;
}

function $sar64(x, n) {
	return x >> $B(n < 64 ? n : 63);
}

function $shlu64(x, n) {
	return n < 64 ? $u64(x << $B(n)) : 0n;
}

function $shru64(x, n) {
	return n < 64 ? x >> $B(n) : 0n;
}

// $scmp compares like Go, which compares the bytes of UTF-8 and thus the code points. JavaScript compares the
// units of UTF-16, which differ only for a surrogate against a unit from U+E000 on.
function $scmp(a, b) {
	if (a === b) return 0;
	const n = $M.min(a.length, b.length);
	for (let i = 0; i < n; i++) {
		const x = a.charCodeAt(i);
		const y = b.charCodeAt(i);
		if (x !== y) {
			const xs = x >= 0xd800 && x <= 0xdfff;
			const ys = y >= 0xd800 && y <= 0xdfff;
			if (xs !== ys) return xs ? 1 : -1;
			return x < y ? -1 : 1;
		}
	}
	return a.length < b.length ? -1 : a.length > b.length ? 1 : 0;
}

// an interface holds the value and the methods of its type
function $box(v, m) {
	return { v: v, m: m };
}

function $ic(i, name, ...args) {
	return i.m[name](i.v, ...args);
}

function $id(v) {
	return v;
}

// list.List is an array

function $lmake(n, zero) {
	const a = [];
	for (let i = 0; i < n; i++) a.push(zero());
	return a;
}

function $lget(a, i, zero, clone) {
	if (i < 0 || i >= a.length) return [zero(), false];
	return [clone(a[i]), true];
}

// $lgetor returns a copy of the element or the default, which the caller has copied already, see List.GetOr
function $lgetor(a, i, d, clone) {
	if (i < 0 || i >= a.length) return d;
	return clone(a[i]);
}

function $lset(a, i, v) {
	if (i < 0 || i >= a.length) return false;
	a[i] = v;
	return true;
}

function $lappend(a, items) {
	for (let i = 0; i < items.length; i++) a.push(items[i]);
}

function $lpop(a, zero) {
	if (a.length === 0) return [zero(), false];
	return [a.pop(), true];
}

function $ldelete(a, i) {
	if (i < 0 || i >= a.length) return false;
	a.splice(i, 1);
	return true;
}

function $lclone(a, clone) {
	return a.map((v) => clone(v));
}

function $lsort(a, cmp) {
	// Array.prototype.sort is stable since ES2019, like slices.SortStableFunc
	a.sort((x, y) => cmp(x, y));
}

// a range over a list visits the elements, which it has, when the range starts, see List.All

function $lall(a, clone) {
	return (y) => {
		for (let i = 0, n = a.length; i < n && i < a.length; i++) {
			if (!y(i, clone(a[i]))) return;
		}
	};
}

function $lvalues(a, clone) {
	return (y) => {
		for (let i = 0, n = a.length; i < n && i < a.length; i++) {
			if (!y(clone(a[i]))) return;
		}
	};
}

// dict.Map is a Map, which keeps the order of insertion like dict.Map of Go

function $dnew() {
	return new Map();
}

function $dget(m, k, zero, clone) {
	if (m.has(k)) return [clone(m.get(k)), true];
	return [zero(), false];
}

// $dgetor returns a copy of the value or the default, which the caller has copied already, see Map.GetOr
function $dgetor(m, k, d, clone) {
	if (m.has(k)) return clone(m.get(k));
	return d;
}

function $dkeys(m) {
	return Array.from(m.keys());
}

function $dall(m, clone) {
	return (y) => {
		const keys = $dkeys(m);
		for (let i = 0; i < keys.length; i++) {
			const k = keys[i];
			if (!m.has(k)) continue;
			if (!y(k, clone(m.get(k)))) return;
		}
	};
}

function $dkeysSeq(m) {
	return (y) => {
		const keys = $dkeys(m);
		for (let i = 0; i < keys.length; i++) {
			if (!m.has(keys[i])) continue;
			if (!y(keys[i])) return;
		}
	};
}

// errors

const $errorString = { Error: (e) => e.s };

// p is the index of the position in the Go source, where the program creates the error, see Bundle.Positions, or
// undefined, e.g. in a binding of a header package
function $errors_New(s, p) {
	return $box({ s: s, $p: p }, $errorString);
}

// math, which behaves like Go also for the special cases

const $math_Abs = $M.abs;
const $math_Acos = $M.acos;
const $math_Asin = $M.asin;
const $math_Atan = $M.atan;
const $math_Atan2 = $M.atan2;
const $math_Cbrt = $M.cbrt;
const $math_Ceil = $M.ceil;
const $math_Cos = $M.cos;
const $math_Exp = $M.exp;
const $math_Floor = $M.floor;
const $math_Hypot = (p, q) => $M.hypot(p, q);
const $math_Log = $M.log;
const $math_Log10 = $M.log10;
const $math_Log2 = $M.log2;
const $math_Sin = $M.sin;
const $math_Sqrt = $M.sqrt;
const $math_Tan = $M.tan;
const $math_Trunc = $M.trunc;
const $math_IsNaN = (f) => f !== f;
const $math_NaN = () => NaN;

// Max is +Inf, if x or y is +Inf, even if the other one is NaN, like math.Max of Go, while Math.max is NaN then.
// Otherwise Math.max agrees with Go: NaN for a NaN and +0 for -0 and +0.
function $math_Max(x, y) {
	if (x === Infinity || y === Infinity) return Infinity;
	return $M.max(x, y);
}

// Min is -Inf, if x or y is -Inf, even if the other one is NaN, like math.Min of Go; otherwise see $math_Max.
function $math_Min(x, y) {
	if (x === -Infinity || y === -Infinity) return -Infinity;
	return $M.min(x, y);
}

function $math_Inf(sign) {
	return sign >= 0 ? Infinity : -Infinity;
}

function $math_IsInf(f, sign) {
	return (sign >= 0 && f === Infinity) || (sign <= 0 && f === -Infinity);
}

function $math_Signbit(x) {
	return x < 0 || (x === 0 && 1 / x < 0);
}

function $math_Copysign(f, sign) {
	return $math_Signbit(f) !== $math_Signbit(sign) ? -f : f;
}

function $math_Mod(x, y) {
	return x % y;
}

function $math_Pow(x, y) {
	if (x === 1) return 1;
	if (x === -1 && (y === Infinity || y === -Infinity)) return 1;
	return $M.pow(x, y);
}

// Round rounds half away from zero, unlike Math.round
function $math_Round(x) {
	const t = $M.trunc(x);
	if ($M.abs(x - t) >= 0.5) return t + $M.sign(x);
	return t;
}

// strings

function $strings_Contains(s, sub) {
	return s.includes(sub);
}

function $strings_HasPrefix(s, p) {
	return s.startsWith(p);
}

function $strings_HasSuffix(s, p) {
	return s.endsWith(p);
}

function $strings_ReplaceAll(s, o, n) {
	if (o === "") return n + Array.from(s).join(n) + (s === "" ? "" : n);
	return s.split(o).join(n);
}

function $strings_TrimPrefix(s, p) {
	return p !== "" && s.startsWith(p) ? s.slice(p.length) : s;
}

function $strings_TrimSuffix(s, p) {
	return p !== "" && s.endsWith(p) ? s.slice(0, s.length - p.length) : s;
}

// the white space of unicode.IsSpace, which differs from \s of JavaScript
function $isSpace(c) {
	return (
		(c >= 9 && c <= 13) ||
		c === 32 ||
		c === 0x85 ||
		c === 0xa0 ||
		c === 0x1680 ||
		(c >= 0x2000 && c <= 0x200a) ||
		c === 0x2028 ||
		c === 0x2029 ||
		c === 0x202f ||
		c === 0x205f ||
		c === 0x3000
	);
}

function $strings_TrimSpace(s) {
	let i = 0;
	let j = s.length;
	while (i < j && $isSpace(s.charCodeAt(i))) i++;
	while (j > i && $isSpace(s.charCodeAt(j - 1))) j--;
	return s.slice(i, j);
}

// text

function $text_RuneCount(s) {
	let n = 0;
	for (let i = 0; i < s.length; i++) {
		const c = s.charCodeAt(i);
		if (c >= 0xd800 && c <= 0xdbff && i + 1 < s.length) i++;
		n++;
	}
	return n;
}

function $text_ByteLen(s) {
	let n = 0;
	for (let i = 0; i < s.length; i++) {
		const c = s.charCodeAt(i);
		if (c < 0x80) n += 1;
		else if (c < 0x800) n += 2;
		else if (c >= 0xd800 && c <= 0xdbff && i + 1 < s.length) {
			n += 4;
			i++;
		} else n += 3;
	}
	return n;
}

function $text_Chars(s) {
	return Array.from(s);
}

function $text_CodePoints(s) {
	return Array.from(s, (c) => c.codePointAt(0));
}

function $text_FromCodePoint(r) {
	if (r < 0 || r > 0x10ffff || (r >= 0xd800 && r <= 0xdfff)) r = 0xfffd;
	return String.fromCodePoint(r);
}

const $text_Compare = $scmp;

function $text_Count(s, sub) {
	if (sub === "") return $text_RuneCount(s) + 1;
	return s.split(sub).length - 1;
}

function $text_Repeat(s, n) {
	return n <= 0 ? "" : s.repeat(n);
}

function $text_Split(s, sep) {
	if (sep === "") return Array.from(s);
	return s.split(sep);
}

function $text_Join(a, sep) {
	return a.join(sep);
}

function $text_Index(s, sub) {
	const i = s.indexOf(sub);
	return i < 0 ? -1 : $text_RuneCount(s.slice(0, i));
}

// Slice joins the elements of $text_Chars from the index from up to the index to, which Array.slice would count
// from the end, if it were negative
function $text_Slice(s, from, to) {
	if (from < 0) from = 0;
	if (from >= to) return "";
	return Array.from(s).slice(from, to).join("");
}

// bin.Bytes is a Uint8Array, a bin.Reader and a bin.Buffer are objects of this runtime

function $bin_fromText(s) {
	const out = [];
	for (const c of s) {
		const p = c.codePointAt(0);
		if (p < 0x80) out.push(p);
		else if (p < 0x800) out.push(0xc0 | (p >> 6), 0x80 | (p & 63));
		else if (p < 0x10000) out.push(0xe0 | (p >> 12), 0x80 | ((p >> 6) & 63), 0x80 | (p & 63));
		else out.push(0xf0 | (p >> 18), 0x80 | ((p >> 12) & 63), 0x80 | ((p >> 6) & 63), 0x80 | (p & 63));
	}
	return Uint8Array.from(out);
}

// $bin_text decodes UTF-8 like Go: a run of invalid bytes becomes one U+FFFD
function $bin_text(b) {
	let s = "";
	let i = 0;
	let bad = false;
	while (i < b.length) {
		const c = b[i];
		if (c < 0x80) {
			s += String.fromCharCode(c);
			i++;
			bad = false;
			continue;
		}
		let n = 0;
		let p = 0;
		let lo = 0x80;
		let hi = 0xbf;
		if (c >= 0xc2 && c <= 0xdf) (n = 1), (p = c & 31);
		else if (c >= 0xe0 && c <= 0xef) {
			n = 2;
			p = c & 15;
			if (c === 0xe0) lo = 0xa0;
			if (c === 0xed) hi = 0x9f;
		} else if (c >= 0xf0 && c <= 0xf4) {
			n = 3;
			p = c & 7;
			if (c === 0xf0) lo = 0x90;
			if (c === 0xf4) hi = 0x8f;
		}
		let ok = n > 0 && i + n < b.length;
		for (let k = 1; ok && k <= n; k++) {
			const d = b[i + k];
			if (d < (k === 1 ? lo : 0x80) || d > (k === 1 ? hi : 0xbf)) ok = false;
			else p = (p << 6) | (d & 63);
		}
		if (ok) {
			s += String.fromCodePoint(p);
			i += n + 1;
			bad = false;
		} else {
			if (!bad) s += "\ufffd";
			bad = true;
			i++;
		}
	}
	return s;
}

function $bin_latin1(b) {
	let s = "";
	for (let i = 0; i < b.length; i++) s += String.fromCharCode(b[i]);
	return s;
}

function $bin_fromLatin1(s) {
	const a = new Uint8Array(s.length);
	let n = 0;
	for (const c of s) {
		const p = c.codePointAt(0);
		if (p > 0xff) return [new Uint8Array(0), false];
		a[n++] = p;
	}
	return [a.subarray(0, n), true];
}

function $bin_hex(b) {
	let s = "";
	for (let i = 0; i < b.length; i++) s += (b[i] < 16 ? "0" : "") + b[i].toString(16);
	return s;
}

function $bin_fromHex(s) {
	if (s.length % 2 !== 0 || !/^[0-9a-fA-F]*$/.test(s)) return [new Uint8Array(0), false];
	const a = new Uint8Array(s.length / 2);
	for (let i = 0; i < a.length; i++) a[i] = parseInt(s.substr(i * 2, 2), 16);
	return [a, true];
}

const $bin_b64 = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/";

function $bin_base64(b) {
	let s = "";
	for (let i = 0; i < b.length; i += 3) {
		const x = (b[i] << 16) | ((b[i + 1] || 0) << 8) | (b[i + 2] || 0);
		s += $bin_b64[(x >> 18) & 63] + $bin_b64[(x >> 12) & 63];
		s += i + 1 < b.length ? $bin_b64[(x >> 6) & 63] : "=";
		s += i + 2 < b.length ? $bin_b64[x & 63] : "=";
	}
	return s;
}

function $bin_fromBase64(s) {
	if (s.length % 4 !== 0 || !/^[A-Za-z0-9+/]*={0,2}$/.test(s)) return [new Uint8Array(0), false];
	const pad = s.endsWith("==") ? 2 : s.endsWith("=") ? 1 : 0;
	const a = new Uint8Array((s.length / 4) * 3 - pad);
	let n = 0;
	for (let i = 0; i < s.length; i += 4) {
		const q = [0, 1, 2, 3].map((k) => (s[i + k] === "=" ? 0 : $bin_b64.indexOf(s[i + k])));
		const x = (q[0] << 18) | (q[1] << 12) | (q[2] << 6) | q[3];
		if (n < a.length) a[n++] = (x >> 16) & 255;
		if (n < a.length) a[n++] = (x >> 8) & 255;
		if (n < a.length) a[n++] = x & 255;
	}
	return [a, true];
}

function $bin_equal(a, b) {
	if (a.length !== b.length) return false;
	for (let i = 0; i < a.length; i++) if (a[i] !== b[i]) return false;
	return true;
}

function $bin_index(b, sub) {
	for (let i = 0; i + sub.length <= b.length; i++) {
		let k = 0;
		while (k < sub.length && b[i + k] === sub[k]) k++;
		if (k === sub.length) return i;
	}
	return -1;
}

function $bin_trimZ(b) {
	const i = b.indexOf(0);
	return i < 0 ? b : b.subarray(0, i);
}

function $bin_open(b) {
	return { b: b, p: 0, bit: 0, e: "", closed: false };
}

function $bin_Decode(data, f, pos) {
	const r = $bin_open(data);
	const v = f(r);
	r.closed = true;
	return [v, r.e !== "" ? $errors_New(r.e, pos) : null];
}

function $bin_failed(r) {
	return r.e !== "" || r.closed;
}

function $bin_fail(r, s) {
	if (r.e === "") r.e = "bin: " + (s === "" ? "invalid data" : s) + " at offset " + r.p;
}

function $bin_align(r) {
	if (r.bit !== 0) {
		r.bit = 0;
		r.p++;
	}
}

// A length, an index or an offset of a program is an int64, whose Number is exact within the length of an array and
// beyond it still out of range, so the functions convert it once with $N and compare it as a Number. An internal
// length is a Number already.

function $bin_in(i, n) {
	const k = $N(i);
	return k >= 0 && k < n;
}

function $bin_slice(b, from, to) {
	const f = $M.max($N(from), 0);
	const t = $M.min($N(to), b.length);
	return f >= t ? b.subarray(0, 0) : b.subarray(f, t);
}

// $bin_take returns the offset of the next n whole bytes or -1, if they do not exist
function $bin_take(r, n) {
	$bin_align(r);
	if ($bin_failed(r)) return -1;
	const k = $N(n);
	if (k < 0 || k > r.b.length - r.p) {
		$bin_fail(r, "reading " + n + " bytes exceeds the length " + r.b.length);
		return -1;
	}
	const at = r.p;
	r.p += k;
	return at;
}

function $bin_bytes(r, n) {
	const at = $bin_take(r, n);
	return at < 0 ? new Uint8Array(0) : r.b.subarray(at, at + $N(n));
}

function $bin_u8(r) {
	const at = $bin_take(r, 1);
	return at < 0 ? 0 : r.b[at];
}

function $bin_un(r, n, le) {
	const at = $bin_take(r, n);
	if (at < 0) return 0;
	let v = 0;
	for (let i = 0; i < n; i++) v = (v * 256 + r.b[le ? at + n - 1 - i : at + i]) >>> 0;
	return v;
}

function $bin_u64(r, le) {
	const at = $bin_take(r, 8);
	if (at < 0) return 0n;
	let v = 0n;
	for (let i = 0; i < 8; i++) v = (v << 8n) | $B(r.b[le ? at + 7 - i : at + i]);
	return v;
}

function $bin_f32(u) {
	const d = new DataView(new ArrayBuffer(4));
	d.setUint32(0, u);
	return d.getFloat32(0);
}

function $bin_f32bits(f) {
	const d = new DataView(new ArrayBuffer(4));
	d.setFloat32(0, f);
	return d.getUint32(0);
}

function $bin_f64(u) {
	const d = new DataView(new ArrayBuffer(8));
	d.setUint32(0, $N(u >> 32n));
	d.setUint32(4, $N(u & 0xffffffffn));
	return d.getFloat64(0);
}

function $bin_f64bits(f) {
	const d = new DataView(new ArrayBuffer(8));
	d.setFloat64(0, f);
	return ($B(d.getUint32(0)) << 32n) | $B(d.getUint32(4));
}

function $bin_checkBits(r, n) {
	if ($bin_failed(r)) return false;
	if (n > 64 || n > (r.b.length - r.p) * 8 - r.bit) {
		$bin_fail(r, "reading " + n + " bits exceeds the data");
		return false;
	}
	return true;
}

function $bin_advance(r, n) {
	r.bit += n;
	if (r.bit === 8) {
		r.bit = 0;
		r.p++;
	}
}

function $bin_bits(r, n) {
	if (!$bin_checkBits(r, n)) return 0n;
	let v = 0n;
	for (let left = n; left > 0; ) {
		const avail = 8 - r.bit;
		const take = $M.min(left, avail);
		const part = (r.b[r.p] >> (avail - take)) & ((1 << take) - 1);
		v = (v << $B(take)) | $B(part);
		$bin_advance(r, take);
		left -= take;
	}
	return v;
}

function $bin_bitsLE(r, n) {
	if (!$bin_checkBits(r, n)) return 0n;
	let v = 0n;
	let shift = 0;
	for (let left = n; left > 0; ) {
		const take = $M.min(left, 8 - r.bit);
		const part = (r.b[r.p] >> r.bit) & ((1 << take) - 1);
		v |= $B(part) << $B(shift);
		shift += take;
		$bin_advance(r, take);
		left -= take;
	}
	return v;
}

function $bin_seek(r, o) {
	if ($bin_failed(r)) return;
	const k = $N(o);
	if (k < 0 || k > r.b.length) {
		$bin_fail(r, "seeking to " + o + " exceeds the length " + r.b.length);
		return;
	}
	r.p = k;
	r.bit = 0;
}

function $bin_pos(r) {
	return r.bit !== 0 ? r.p + 1 : r.p;
}

function $bin_remaining(r) {
	return $bin_failed(r) ? 0 : r.b.length - $bin_pos(r);
}

// A buffer has no limit in the browser. The capacity is a hint, of which it reserves at most 64 MiB in advance.
const $bin_maxReserve = 67108864;

function $bin_newBuffer(c) {
	return { a: new Uint8Array($M.max($M.min($N(c), $bin_maxReserve), 16)), n: 0, bit: 0 };
}

function $bin_reserve(b, n) {
	b.bit = 0;
	if (b.n + n > b.a.length) {
		const a = new Uint8Array($M.max(b.a.length * 2, b.n + n));
		a.set(b.a.subarray(0, b.n));
		b.a = a;
	}
	const at = b.n;
	b.n += n;
	return at;
}

function $bin_put(a, at, n, le, v) {
	for (let i = 0; i < n; i++) a[le ? at + i : at + n - 1 - i] = (v / $M.pow(256, i)) & 255;
}

function $bin_put64(a, at, le, v) {
	for (let i = 0; i < 8; i++) a[le ? at + i : at + 7 - i] = $N((v >> $B(i * 8)) & 255n);
}

// $bin_reserve may replace b.a, so the callers read b.a after it

function $bin_append(b, n, le, v) {
	const at = $bin_reserve(b, n);
	$bin_put(b.a, at, n, le, v);
}

function $bin_append64(b, le, v) {
	const at = $bin_reserve(b, 8);
	$bin_put64(b.a, at, le, v);
}

function $bin_set(b, i, n, le, v) {
	const k = $N(i);
	if (k < 0 || k > b.n - n) return false;
	$bin_put(b.a, k, n, le, v);
	return true;
}

function $bin_set64(b, i, le, v) {
	const k = $N(i);
	if (k < 0 || k > b.n - 8) return false;
	$bin_put64(b.a, k, le, v);
	return true;
}

function $bin_appendBytes(b, p) {
	const at = $bin_reserve(b, p.length);
	b.a.set(p, at);
}

function $bin_appendBit(b, v, msb) {
	if (b.bit === 0) $bin_reserve(b, 1);
	if (v) b.a[b.n - 1] |= msb ? 1 << (7 - b.bit) : 1 << b.bit;
	b.bit = (b.bit + 1) % 8;
}

function $bin_appendBits(b, v, n, msb) {
	n = $M.min(n, 64);
	for (let i = 0; i < n; i++) {
		const k = msb ? n - 1 - i : i;
		$bin_appendBit(b, ((v >> $B(k)) & 1n) === 1n, msb);
	}
}

function $bin_reset(b) {
	b.n = 0;
	b.bit = 0;
}

function $bin_freeze(b) {
	const res = b.a.subarray(0, b.n);
	b.a = new Uint8Array(16);
	$bin_reset(b);
	return [res, null];
}

function $bin_rot64(x, k) {
	return k === 0 ? x : $u64((x << $B(k)) | (x >> $B(64 - k)));
}

function $bin_pop32(x) {
	let n = 0;
	for (let v = x >>> 0; v !== 0; v = (v & (v - 1)) >>> 0) n++;
	return n;
}

function $bin_tz64(x) {
	if (x === 0n) return 64;
	const lo = $N(x & 0xffffffffn);
	if (lo !== 0) return 31 - $M.clz32(lo & -lo);
	const hi = $N(x >> 32n);
	return 32 + 31 - $M.clz32(hi & -hi);
}

function $bin_swap32(x) {
	return (((x & 0xff) << 24) | ((x & 0xff00) << 8) | ((x >>> 8) & 0xff00) | (x >>> 24)) >>> 0;
}

function $bin_swap64(x) {
	return ($B($bin_swap32($N(x & 0xffffffffn))) << 32n) | $B($bin_swap32($N(x >> 32n)));
}

function $bin_sext32(v, n) {
	if (n === 0) return 0;
	if (n >= 32) return v | 0;
	const s = 32 - n;
	return (v << s) >> s;
}

// opt.Option is null or an object with the value, which never changes, so a copy of it is shared

function $oget(o, zero, clone) {
	if (o === null) return [zero(), false];
	return [clone(o.v), true];
}

function $oor(o, d, clone) {
	return o === null ? d : clone(o.v);
}

// $ozero returns v or the zero value, unless ok, see opt.OrZero; v is a copy already, e.g. the result of $lget
function $ozero(zero, v, ok) {
	return ok ? v : zero();
}

// $ozeroerr returns v or the zero value, if there is an error, see opt.OrZeroErr
function $ozeroerr(zero, v, err) {
	return err === null ? v : zero();
}

function $text_FromInt64(n) {
	return n.toString();
}

function $text_FromUint64(n) {
	return n.toString();
}

// $parseLong parses the digits and the sign without BigInt, whose parser differs for a sign between the engines
function $parseLong(s, fn, min, max, p) {
	const neg = s[0] === "-";
	const d = neg || s[0] === "+" ? s.slice(1) : s;
	if (d === "" || !/^[0-9]+$/.test(d) || (neg && min === 0n)) return [0n, $errors_New("text." + fn + ": invalid syntax: " + s, p)];
	let v = 0n;
	for (let i = 0; i < d.length; i++) {
		v = v * 10n + $B(d.charCodeAt(i) - 48);
		if (v > 18446744073709551616n) break;
	}
	if (neg) v = -v;
	if (v < min || v > max) return [0n, $errors_New("text." + fn + ": value out of range: " + s, p)];
	return [v, null];
}

function $text_ParseInt64(s, p) {
	return $parseLong(s, "ParseInt64", -9223372036854775808n, 9223372036854775807n, p);
}

function $text_ParseUint64(s, p) {
	return $parseLong(s, "ParseUint64", 0n, 18446744073709551615n, p);
}

function $text_FromInt(n) {
	return "" + n;
}

// $text_FromFloat formats like strconv.FormatFloat(f, 'f', d, 64), which rounds the exact binary value half to
// even. toFixed rounds the exact value too, but a tie away from zero, so a tie, which has become odd, is corrected.
function $text_FromFloat(f, d) {
	d = $M.min($M.max(d, 0), 100);
	if (f !== f) return "NaN";
	if (f === Infinity) return "+Inf";
	if (f === -Infinity) return "-Inf";
	const neg = f < 0 || Object.is(f, -0);
	const a = $M.abs(f);
	let s;
	if (a >= 1e21) {
		// an integer, which toFixed would write with an exponent
		s = BigInt(a).toString() + (d > 0 ? "." + "0".repeat(d) : "");
	} else {
		s = a.toFixed(d);
		if ($ftie(a, d)) {
			const digits = s.replace(".", "");
			if ((digits.charCodeAt(digits.length - 1) - 48) % 2 === 1) {
				const n = (BigInt(digits) - 1n).toString().padStart(digits.length, "0");
				s = d > 0 ? n.slice(0, n.length - d) + "." + n.slice(n.length - d) : n;
			}
		}
	}
	return neg ? "-" + s : s;
}

// $ftie tells, whether the positive finite a lies exactly in the middle between two numbers with d decimals: a is
// m * 2^e, so 2 * a * 10^d = m * 5^d * 2^(e+1+d) is an odd integer, if the trailing zeros of m cancel the power.
function $ftie(a, d) {
	if (a === 0) return false;
	const v = new DataView(new ArrayBuffer(8));
	v.setFloat64(0, a);
	const hi = v.getUint32(0);
	const lo = v.getUint32(4);
	let e = (hi >>> 20) & 0x7ff;
	let m = (BigInt(hi & 0xfffff) << 32n) | BigInt(lo);
	if (e === 0) {
		e = 1;
	} else {
		m |= 1n << 52n;
	}
	e -= 1075;
	let tz = 0;
	while (((m >> BigInt(tz)) & 1n) === 0n) tz++;
	return tz + e + 1 + d === 0;
}

function $isDigit(c) {
	return c >= 48 && c <= 57;
}

function $text_GroupDigits(s, dm, gm) {
	const n = s.length;
	const i = n > 0 && (s[0] === "+" || s[0] === "-") ? 1 : 0;
	let j = i;
	while (j < n && $isDigit(s.charCodeAt(j))) j++;
	if (j === i) return s;
	if (j < n) {
		if (s[j] !== "." || j + 1 === n) return s;
		for (let k = j + 1; k < n; k++) if (!$isDigit(s.charCodeAt(k))) return s;
	}
	let r = s.slice(0, i);
	for (let k = i; k < j; k++) {
		if (k > i && (j - k) % 3 === 0) r += gm;
		r += s[k];
	}
	return j < n ? r + dm + s.slice(j + 1) : r;
}

function $text_ParseInt(s, p) {
	let d = s;
	let neg = false;
	if (d.length > 0 && (d[0] === "+" || d[0] === "-")) {
		neg = d[0] === "-";
		d = d.slice(1);
	}
	if (d === "") return [0, $errors_New("text.ParseInt: invalid syntax: " + s, p)];
	let n = 0;
	for (let i = 0; i < d.length; i++) {
		const c = d.charCodeAt(i);
		if (c < 48 || c > 57) return [0, $errors_New("text.ParseInt: invalid syntax: " + s, p)];
		n = n * 10 + (c - 48);
		if (n > 2147483648) return [0, $errors_New("text.ParseInt: value out of range: " + s, p)];
	}
	if (neg) n = 0 - n;
	if (n > 2147483647) return [0, $errors_New("text.ParseInt: value out of range: " + s, p)];
	return [n | 0, null];
}

// $text_ParseFloat rounds like strconv.ParseFloat of Go, for any number of digits, without the parsing of
// JavaScript: it scans the form like text.ParseFloat into the digits without the zeros at both ends and the
// exponent, so that the magnitude is 0.digits * 10^exp.
function $text_ParseFloat(s, p) {
	const n = s.length;
	let i = 0;
	let neg = false;
	if (i < n && (s[i] === "+" || s[i] === "-")) {
		neg = s[i] === "-";
		i++;
	}
	let d = "";
	let exp = 0;
	let seen = false;
	let point = false;
	for (; i < n; i++) {
		const c = s.charCodeAt(i);
		if (c === 46 && !point) {
			point = true;
			continue;
		}
		if (!$isDigit(c)) break;
		seen = true;
		if (c === 48 && d === "") {
			// a leading zero after the point moves the digits to the right
			if (point) exp--;
		} else {
			d += s[i];
			if (!point) exp++;
		}
	}
	if (seen && i < n && (s[i] === "e" || s[i] === "E")) {
		i++;
		let eneg = false;
		if (i < n && (s[i] === "+" || s[i] === "-")) {
			eneg = s[i] === "-";
			i++;
		}
		// the exponent stops growing at 1e12, like in Go: any number of digits is far less, so a number with a
		// greater exponent is out of range or 0 anyway
		const start = i;
		let e = 0;
		for (; i < n && $isDigit(s.charCodeAt(i)); i++) {
			if (e < 1e12) e = e * 10 + (s.charCodeAt(i) - 48);
		}
		if (i === start) seen = false;
		exp += eneg ? -e : e;
	}
	if (!seen || i < n) return [0, $errors_New("text.ParseFloat: invalid syntax: " + s, p)];
	let k = d.length;
	while (k > 0 && d.charCodeAt(k - 1) === 48) k--;
	d = d.slice(0, k);
	if (d === "" || exp < -323) {
		// 0.digits * 10^exp is less than 10^-324, which is less than half of the least subnormal number
		return [neg ? -0 : 0, null];
	}
	// from 10^309 on, a number is beyond the greatest float64
	const f = exp > 309 ? Infinity : $pfloat(d, exp);
	if (f === Infinity) return [0, $errors_New("text.ParseFloat: value out of range: " + s, p)];
	return [neg ? -f : f, null];
}

// the powers of ten, which are exact float64 numbers
const $pow10 = [
	1, 1e1, 1e2, 1e3, 1e4, 1e5, 1e6, 1e7, 1e8, 1e9, 1e10, 1e11,
	1e12, 1e13, 1e14, 1e15, 1e16, 1e17, 1e18, 1e19, 1e20, 1e21, 1e22,
];

// $pfloat returns the float64 nearest to 0.d * 10^exp, ties to even, or Infinity. With at most 15 digits and a
// power of ten up to 10^22, both are exact float64 numbers, so a single multiplication or division rounds
// correctly. Otherwise it divides exactly as BigInt. A float64 or a midpoint between two has at most 768
// significant digits, so 800 digits and a 1 for any further ones round the same.
function $pfloat(d, exp) {
	let q = exp - d.length;
	if (d.length <= 15 && q >= -22 && q <= 22) {
		let m = 0;
		for (let i = 0; i < d.length; i++) m = m * 10 + (d.charCodeAt(i) - 48);
		return q >= 0 ? m * $pow10[q] : m / $pow10[-q];
	}
	if (d.length > 800) {
		d = d.slice(0, 800) + "1";
		q = exp - 801;
	}
	let num = BigInt(d);
	let den = 1n;
	if (q >= 0) num *= 10n ** BigInt(q);
	else den = 10n ** BigInt(-q);
	return $pfround(num, den);
}

// $pfround returns the float64 nearest to num / den, ties to even, or Infinity: it is m * 2^k with an integer m of
// 53 bits, or of fewer bits for a subnormal number, whose k is -1074.
function $pfround(num, den) {
	// with a bits of num and b bits of den, num / den lies between 2^(a-b-1) and 2^(a-b+1), so m has 53 or 54
	// bits, or fewer at the least k
	let k = num.toString(2).length - den.toString(2).length - 53;
	if (k < -1074) k = -1074;
	let [m, r, dk] = $pfdiv(num, den, k);
	if (m >= 9007199254740992n) {
		k++;
		[m, r, dk] = $pfdiv(num, den, k);
	}
	if (2n * r > dk || (2n * r === dk && (m & 1n) === 1n)) m++;
	if (m === 9007199254740992n) {
		m = 4503599627370496n;
		k++;
	}
	if (k > 971) return Infinity;
	// a normal number is (k + 1075) << 52 | (m - 2^52), which is (k + 1074) << 52 + m, and a subnormal one is m
	const v = new DataView(new ArrayBuffer(8));
	v.setBigUint64(0, (BigInt(k + 1074) << 52n) + m);
	return v.getFloat64(0);
}

// $pfdiv returns the quotient and the remainder of num / (den * 2^k) and the divisor, by which it divided
function $pfdiv(num, den, k) {
	let dk = den;
	if (k >= 0) dk = den << BigInt(k);
	else num = num << BigInt(-k);
	const m = num / dk;
	return [m, num - m * dk, dk];
}

function $text_UpperASCII(s) {
	let r = "";
	for (let i = 0; i < s.length; i++) {
		const c = s.charCodeAt(i);
		r += c >= 97 && c <= 122 ? String.fromCharCode(c - 32) : s[i];
	}
	return r;
}

function $text_LowerASCII(s) {
	let r = "";
	for (let i = 0; i < s.length; i++) {
		const c = s.charCodeAt(i);
		r += c >= 65 && c <= 90 ? String.fromCharCode(c + 32) : s[i];
	}
	return r;
}

// num

function $num_Trunc64(f) {
	if (f !== f) return 0n;
	if (f >= 9223372036854775807) return 9223372036854775807n;
	if (f <= -9223372036854775808) return -9223372036854775808n;
	return $B($M.trunc(f));
}

function $num_TruncU64(f) {
	if (f !== f || f <= 0) return 0n;
	if (f >= 18446744073709551615) return 18446744073709551615n;
	return $B($M.trunc(f));
}

function $num_Round64(f) {
	return $num_Trunc64($math_Round(f));
}

function $num_Trunc32(f) {
	if (f !== f) return 0;
	if (f >= 2147483647) return 2147483647;
	if (f <= -2147483648) return -2147483648;
	return $M.trunc(f) | 0;
}

function $num_TruncU32(f) {
	if (f !== f || f <= 0) return 0;
	if (f >= 4294967295) return 4294967295;
	return $M.trunc(f) >>> 0;
}

function $num_Round32(f) {
	return $num_Trunc32($math_Round(f));
}

// json.Decode reads like package std/json/core: the same documents, the same errors at the same offsets. A type is a
// function, which returns its description, see emitter.jsonType.

function $json_Decode(data, td, pos) {
	const d = { b: data, p: 0, depth: 0, e: "" };
	const v = $j_value(d, td);
	$j_end(d);
	if (d.e !== "") return [$j_empty(td()), $errors_New(d.e, pos)];
	return [v, null];
}

function $j_value(d, td) {
	const t = td();
	if ($j_null(d)) return $j_empty(t);
	switch (t.k) {
		case "b":
			return $j_bool(d);
		case "i":
			return Number($j_integer(d, -2147483648n, 2147483647n));
		case "u":
			return Number($j_integer(d, 0n, 4294967295n));
		case "I":
			return $j_integer(d, -9223372036854775808n, 9223372036854775807n);
		case "U":
			return $j_integer(d, 0n, 18446744073709551615n);
		case "f":
			return $j_float(d);
		case "s": {
			const s = $j_str(d, "a string");
			return s === null ? "" : s;
		}
		case "y":
			return $j_base64(d);
		case "n":
			return $j_big(d);
		case "o":
			return { v: $j_value(d, t.e) };
		case "l": {
			const a = [];
			if ($j_open(d, 91, "an array")) for (let i = 0; $j_next(d, i, 93); i++) a.push($j_value(d, t.e));
			return a;
		}
		case "m": {
			const m = $dnew();
			if ($j_open(d, 123, "an object"))
				for (let i = 0; $j_next(d, i, 125); i++) {
					const k = $j_key(d);
					if (k === null) break;
					m.set(k, $j_value(d, t.e));
				}
			return m;
		}
		case "r": {
			const o = t.z();
			if ($j_open(d, 123, "an object"))
				for (let i = 0; $j_next(d, i, 125); i++) {
					const k = $j_key(d);
					if (k === null) break;
					const f = t.f.get(k);
					if (f === undefined) $j_skip(d);
					else o[f[0]] = $j_value(d, f[1]);
				}
			return o;
		}
	}
	throw new Error("json: unknown type " + t.k);
}

function $j_empty(t) {
	switch (t.k) {
		case "b":
			return false;
		case "i":
		case "u":
		case "f":
			return 0;
		case "I":
		case "U":
		case "n":
			return 0n;
		case "s":
			return "";
		case "y":
			return new Uint8Array(0);
		case "o":
			return null;
		case "l":
			return [];
		case "m":
			return $dnew();
	}
	return t.z();
}

function $j_fail(d, p, msg) {
	if (d.e === "") d.e = "json: " + msg + " at offset " + p;
}

function $j_ws(d) {
	const b = d.b;
	while (d.p < b.length) {
		const c = b[d.p];
		if (c !== 32 && c !== 9 && c !== 10 && c !== 13) return;
		d.p++;
	}
}

function $j_unexpected(d) {
	if (d.p >= d.b.length) {
		$j_fail(d, d.p, "unexpected end of input");
		return;
	}
	const c = d.b[d.p];
	if (c >= 0x20 && c < 0x7f) $j_fail(d, d.p, "unexpected character '" + String.fromCharCode(c) + "'");
	else $j_fail(d, d.p, "unexpected byte 0x" + (c < 16 ? "0" : "") + c.toString(16));
}

function $j_expect(d, what) {
	const c = d.p < d.b.length ? d.b[d.p] : -1;
	const starts = c === 34 || c === 123 || c === 91 || c === 116 || c === 102 || c === 110 || c === 45 || (c >= 48 && c <= 57);
	if (starts) $j_fail(d, d.p, "expected " + what);
	else $j_unexpected(d);
}

function $j_lit(d, lit) {
	for (let i = 0; i < lit.length; i++) {
		if (d.p >= d.b.length || d.b[d.p] !== lit.charCodeAt(i)) {
			$j_unexpected(d);
			return;
		}
		d.p++;
	}
}

function $j_null(d) {
	if (d.e !== "") return true;
	$j_ws(d);
	if (d.p < d.b.length && d.b[d.p] === 110) {
		$j_lit(d, "null");
		return true;
	}
	return false;
}

function $j_bool(d) {
	if (d.e !== "") return false;
	$j_ws(d);
	const c = d.p < d.b.length ? d.b[d.p] : -1;
	if (c === 116) {
		$j_lit(d, "true");
		return d.e === "";
	}
	if (c === 102) $j_lit(d, "false");
	else $j_expect(d, "a boolean");
	return false;
}

function $j_digit(d) {
	return d.p < d.b.length && d.b[d.p] >= 48 && d.b[d.p] <= 57;
}

// $j_number reads a number and returns its start, its end and whether it is an integer, or null
function $j_number(d, what) {
	$j_ws(d);
	const b = d.b;
	const start = d.p;
	if (d.p >= b.length || (b[d.p] !== 45 && (b[d.p] < 48 || b[d.p] > 57))) {
		$j_expect(d, what);
		return null;
	}
	if (b[d.p] === 45) d.p++;
	if (d.p < b.length && b[d.p] === 48) d.p++;
	else if (d.p < b.length && b[d.p] >= 49 && b[d.p] <= 57) while ($j_digit(d)) d.p++;
	else {
		$j_unexpected(d);
		return null;
	}
	let integer = true;
	if (d.p < b.length && b[d.p] === 46) {
		integer = false;
		d.p++;
		if (!$j_digit(d)) {
			$j_unexpected(d);
			return null;
		}
		while ($j_digit(d)) d.p++;
	}
	if (d.p < b.length && (b[d.p] === 101 || b[d.p] === 69)) {
		integer = false;
		d.p++;
		if (d.p < b.length && (b[d.p] === 43 || b[d.p] === 45)) d.p++;
		if (!$j_digit(d)) {
			$j_unexpected(d);
			return null;
		}
		while ($j_digit(d)) d.p++;
	}
	return [start, d.p, integer];
}

function $j_ascii(b, s, e) {
	let r = "";
	for (let i = s; i < e; i++) r += String.fromCharCode(b[i]);
	return r;
}

// $j_integer reads an integer within min and max as a BigInt
function $j_integer(d, min, max) {
	if (d.e !== "") return 0n;
	const n = $j_number(d, "an integer");
	if (n === null || d.e !== "") return 0n;
	if (!n[2]) {
		$j_fail(d, n[0], "expected an integer");
		return 0n;
	}
	const v = BigInt($j_ascii(d.b, n[0], n[1]));
	if (v < min || v > max) {
		$j_fail(d, n[0], "number out of range");
		return 0n;
	}
	return v;
}

function $j_float(d) {
	if (d.e !== "") return 0;
	const n = $j_number(d, "a number");
	if (n === null || d.e !== "") return 0;
	const f = Number($j_ascii(d.b, n[0], n[1]));
	if (f === Infinity || f === -Infinity) {
		$j_fail(d, n[0], "number out of range");
		return 0;
	}
	return f;
}

function $j_big(d) {
	if (d.e !== "") return 0n;
	const n = $j_number(d, "an integer");
	if (n === null || d.e !== "") return 0n;
	if (!n[2]) {
		$j_fail(d, n[0], "expected an integer");
		return 0n;
	}
	return BigInt($j_ascii(d.b, n[0], n[1]));
}

// $j_str reads a string or returns null. Each segment between escapes is decoded like bin.Bytes.Text.
function $j_str(d, what) {
	if (d.e !== "") return null;
	$j_ws(d);
	const b = d.b;
	if (d.p >= b.length || b[d.p] !== 34) {
		$j_expect(d, what);
		return null;
	}
	d.p++;
	let out = "";
	let seg = d.p;
	for (;;) {
		if (d.p >= b.length) {
			$j_fail(d, d.p, "unexpected end of input");
			return null;
		}
		const c = b[d.p];
		if (c === 34) {
			if (seg < d.p) out += $bin_text(b.subarray(seg, d.p));
			d.p++;
			return out;
		}
		if (c < 0x20) {
			$j_fail(d, d.p, "invalid character in string");
			return null;
		}
		if (c === 92) {
			if (seg < d.p) out += $bin_text(b.subarray(seg, d.p));
			const r = $j_escape(d);
			if (r === null) return null;
			out += r;
			seg = d.p;
			continue;
		}
		d.p++;
	}
}

function $j_escape(d) {
	const b = d.b;
	const at = d.p;
	d.p++;
	if (d.p >= b.length) {
		$j_fail(d, d.p, "unexpected end of input");
		return null;
	}
	const c = b[d.p++];
	switch (c) {
		case 34:
			return '"';
		case 92:
			return "\\";
		case 47:
			return "/";
		case 98:
			return "\b";
		case 102:
			return "\f";
		case 110:
			return "\n";
		case 114:
			return "\r";
		case 116:
			return "\t";
		case 117: {
			const r = $j_hex4(d);
			if (r < 0) break;
			if (r >= 0xd800 && r < 0xdc00) {
				// a high surrogate needs its low surrogate in the next escape, otherwise it is invalid on its own
				if (d.p + 1 < b.length && b[d.p] === 92 && b[d.p + 1] === 117) {
					const save = d.p;
					d.p += 2;
					const lo = $j_hex4(d);
					if (lo >= 0xdc00 && lo < 0xe000) return String.fromCharCode(r, lo);
					d.p = save;
				}
				return "\ufffd";
			}
			if (r >= 0xdc00 && r < 0xe000) return "\ufffd";
			return String.fromCharCode(r);
		}
	}
	$j_fail(d, at, "invalid escape");
	return null;
}

function $j_hex4(d) {
	if (d.p + 4 > d.b.length) return -1;
	let r = 0;
	for (let i = 0; i < 4; i++) {
		const c = d.b[d.p + i];
		let v;
		if (c >= 48 && c <= 57) v = c - 48;
		else if (c >= 97 && c <= 102) v = c - 87;
		else if (c >= 65 && c <= 70) v = c - 55;
		else return -1;
		r = r * 16 + v;
	}
	d.p += 4;
	return r;
}

function $j_base64(d) {
	if (d.e !== "") return new Uint8Array(0);
	$j_ws(d);
	const start = d.p;
	const s = $j_str(d, "a string");
	if (s === null) return new Uint8Array(0);
	const r = $bin_fromBase64(s);
	if (!r[1]) {
		$j_fail(d, start, "invalid base64");
		return new Uint8Array(0);
	}
	return r[0];
}

function $j_open(d, c, what) {
	if (d.e !== "") return false;
	$j_ws(d);
	if (d.p >= d.b.length || d.b[d.p] !== c) {
		$j_expect(d, what);
		return false;
	}
	if (d.depth === 512) {
		$j_fail(d, d.p, "nesting too deep");
		return false;
	}
	d.depth++;
	d.p++;
	return true;
}

// $j_next tells, whether the array or object has its element i, and reads the comma before it or the end after the
// last one
function $j_next(d, i, end) {
	if (d.e !== "") return false;
	$j_ws(d);
	const b = d.b;
	if (i === 0 && d.p < b.length && b[d.p] === end) {
		d.p++;
		d.depth--;
		return false;
	}
	if (i > 0) {
		if (d.p < b.length && b[d.p] === 44) d.p++;
		else if (d.p < b.length && b[d.p] === end) {
			d.p++;
			d.depth--;
			return false;
		} else {
			$j_unexpected(d);
			return false;
		}
	}
	return true;
}

function $j_key(d) {
	const k = $j_str(d, "a string");
	if (k === null) return null;
	$j_ws(d);
	if (d.p >= d.b.length || d.b[d.p] !== 58) {
		$j_unexpected(d);
		return null;
	}
	d.p++;
	return k;
}

function $j_skip(d) {
	if (d.e !== "") return;
	$j_ws(d);
	if (d.p >= d.b.length) {
		$j_unexpected(d);
		return;
	}
	const c = d.b[d.p];
	if (c === 34) $j_str(d, "a string");
	else if (c === 123) {
		if ($j_open(d, 123, "an object"))
			for (let i = 0; $j_next(d, i, 125); i++) {
				if ($j_key(d) === null) return;
				$j_skip(d);
			}
	} else if (c === 91) {
		if ($j_open(d, 91, "an array")) for (let i = 0; $j_next(d, i, 93); i++) $j_skip(d);
	} else if (c === 116) $j_lit(d, "true");
	else if (c === 102) $j_lit(d, "false");
	else if (c === 110) $j_lit(d, "null");
	else $j_number(d, "a value");
}

function $j_end(d) {
	if (d.e !== "") return;
	$j_ws(d);
	if (d.p < d.b.length) $j_fail(d, d.p, "unexpected data after the value");
}

// json.Encode writes like package std/json/core: the same bytes in every engine.

function $json_Encode(v, td, pos) {
	const e = { o: [], e: "" };
	$je_value(e, v, td);
	if (e.e !== "") return [new Uint8Array(0), $errors_New(e.e, pos)];
	return [$bin_fromText(e.o.join("")), null];
}

function $je_value(e, v, td) {
	const t = td();
	switch (t.k) {
		case "b":
			e.o.push(v ? "true" : "false");
			return;
		case "i":
		case "u":
			e.o.push(v === 0 ? "0" : String(v));
			return;
		case "I":
		case "U":
		case "n":
			e.o.push(v.toString());
			return;
		case "f":
			$je_float(e, v);
			return;
		case "s":
			$je_string(e, v);
			return;
		case "y":
			e.o.push('"' + $bin_base64(v) + '"');
			return;
		case "o":
			if (v === null) e.o.push("null");
			else $je_value(e, v.v, t.e);
			return;
		case "l":
			e.o.push("[");
			for (let i = 0; i < v.length; i++) {
				if (i > 0) e.o.push(",");
				$je_value(e, v[i], t.e);
			}
			e.o.push("]");
			return;
		case "m": {
			e.o.push("{");
			let i = 0;
			for (const [k, x] of v) {
				if (i++ > 0) e.o.push(",");
				$je_string(e, k);
				e.o.push(":");
				$je_value(e, x, t.e);
			}
			e.o.push("}");
			return;
		}
		case "r": {
			e.o.push("{");
			let first = true;
			for (const [k, f] of t.f) {
				const mark = e.o.length;
				if (!first) e.o.push(",");
				$je_string(e, k);
				e.o.push(":");
				const start = e.o.length;
				$je_value(e, v[f[0]], f[1]);
				if (f[2]) {
					const w = e.o.slice(start).join("");
					if (w === "null" || w === '""' || w === "[]" || w === "{}") {
						e.o.length = mark;
						continue;
					}
				}
				first = false;
			}
			e.o.push("}");
			return;
		}
	}
	throw new Error("json: unknown type " + t.k);
}

// $je_float writes like Number.prototype.toString, but -0 as -0
function $je_float(e, f) {
	if (f !== f || f === Infinity || f === -Infinity) {
		if (e.e === "") e.e = "json: unsupported value " + (f !== f ? "NaN" : f > 0 ? "+Inf" : "-Inf");
		e.o.push("null");
		return;
	}
	if (f === 0) e.o.push(1 / f < 0 ? "-0" : "0");
	else e.o.push(String(f));
}

// $je_string escapes only the quote, the backslash and the control characters, and writes a lone surrogate as U+FFFD
function $je_string(e, s) {
	let out = '"';
	let start = 0;
	for (let i = 0; i < s.length; i++) {
		const c = s.charCodeAt(i);
		if (c >= 0x20 && c !== 34 && c !== 92 && (c < 0xd800 || c > 0xdfff)) continue;
		if (c >= 0xd800 && c < 0xdc00 && i + 1 < s.length) {
			const lo = s.charCodeAt(i + 1);
			if (lo >= 0xdc00 && lo <= 0xdfff) {
				i++;
				continue;
			}
		}
		out += s.slice(start, i);
		if (c === 34) out += '\\"';
		else if (c === 92) out += "\\\\";
		else if (c === 8) out += "\\b";
		else if (c === 12) out += "\\f";
		else if (c === 10) out += "\\n";
		else if (c === 13) out += "\\r";
		else if (c === 9) out += "\\t";
		else if (c < 0x20) out += "\\u00" + (c < 16 ? "0" : "") + c.toString(16);
		else out += "�";
		start = i + 1;
	}
	e.o.push(out + s.slice(start) + '"');
}

