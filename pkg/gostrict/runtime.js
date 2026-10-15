// The runtime of the gostrict subset. Every name starts with $, which a Go identifier can not contain, so the
// translated code never collides with it. The translated code refers to the globals of JavaScript only through
// this runtime, so a Go variable may be called Math or Map.
//
// The checker proves, that a program can not fail at runtime: there is no nil, no index out of range and no
// division by zero. So the runtime contains no checks, only the semantics of Go, which JavaScript lacks.

const $M = Math;

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

function $lall(a, clone) {
	return (y) => {
		for (let i = 0; i < a.length; i++) {
			if (!y(i, clone(a[i]))) return;
		}
	};
}

function $lvalues(a, clone) {
	return (y) => {
		for (let i = 0; i < a.length; i++) {
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

function $errors_New(s) {
	return $box({ s: s }, $errorString);
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
const $math_Max = (x, y) => $M.max(x, y);
const $math_Min = (x, y) => $M.min(x, y);
const $math_Sin = $M.sin;
const $math_Sqrt = $M.sqrt;
const $math_Tan = $M.tan;
const $math_Trunc = $M.trunc;
const $math_IsNaN = (f) => f !== f;
const $math_NaN = () => NaN;

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

// opt.Option is null or an object with the value, which never changes, so a copy of it is shared

function $oget(o, zero, clone) {
	if (o === null) return [zero(), false];
	return [clone(o.v), true];
}

function $oor(o, d, clone) {
	return o === null ? d : clone(o.v);
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

function $text_ParseInt(s) {
	let d = s;
	let neg = false;
	if (d.length > 0 && (d[0] === "+" || d[0] === "-")) {
		neg = d[0] === "-";
		d = d.slice(1);
	}
	if (d === "") return [0, $errors_New("text.ParseInt: invalid syntax: " + s)];
	let n = 0;
	for (let i = 0; i < d.length; i++) {
		const c = d.charCodeAt(i);
		if (c < 48 || c > 57) return [0, $errors_New("text.ParseInt: invalid syntax: " + s)];
		n = n * 10 + (c - 48);
		if (n > 2147483648) return [0, $errors_New("text.ParseInt: value out of range: " + s)];
	}
	if (neg) n = 0 - n;
	if (n > 2147483647) return [0, $errors_New("text.ParseInt: value out of range: " + s)];
	return [n | 0, null];
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
