type QrVersionSpec = {
	version: number;
	size: number;
	dataCodewords: number;
	errorCodewords: number;
	alignmentCenter: number | null;
};

const VERSION_SPECS: QrVersionSpec[] = [
	{ version: 1, size: 21, dataCodewords: 19, errorCodewords: 7, alignmentCenter: null },
	{ version: 2, size: 25, dataCodewords: 34, errorCodewords: 10, alignmentCenter: 18 },
	{ version: 3, size: 29, dataCodewords: 55, errorCodewords: 15, alignmentCenter: 22 },
	{ version: 4, size: 33, dataCodewords: 80, errorCodewords: 20, alignmentCenter: 26 },
	{ version: 5, size: 37, dataCodewords: 108, errorCodewords: 26, alignmentCenter: 30 },
];

const FORMAT_BITS_LEVEL_L_MASK_0 = 0x77c4;

export function qrCodeDataUrl(text: string): string | null {
	const dataBytes = Array.from(new TextEncoder().encode(text));
	const spec = VERSION_SPECS.find((item) => dataBytes.length + 2 <= item.dataCodewords);
	if (!spec) return null;

	const dataCodewords = encodeDataCodewords(dataBytes, spec.dataCodewords);
	const errorCodewords = computeErrorCodewords(dataCodewords, spec.errorCodewords);
	const codewords = [...dataCodewords, ...errorCodewords];
	const modules = buildModules(spec, codewords);
	return svgDataUrl(modules);
}

function encodeDataCodewords(dataBytes: number[], capacity: number) {
	const bits: number[] = [];
	appendBits(bits, 0b0100, 4);
	appendBits(bits, dataBytes.length, 8);
	for (const byte of dataBytes) appendBits(bits, byte, 8);

	const maxBits = capacity * 8;
	appendBits(bits, 0, Math.min(4, maxBits - bits.length));
	while (bits.length % 8 !== 0) bits.push(0);

	const codewords: number[] = [];
	for (let bitIndex = 0; bitIndex < bits.length; bitIndex += 8) {
		let codeword = 0;
		for (let offset = 0; offset < 8; offset += 1) {
			codeword = (codeword << 1) | bits[bitIndex + offset];
		}
		codewords.push(codeword);
	}

	for (let padIndex = 0; codewords.length < capacity; padIndex += 1) {
		codewords.push(padIndex % 2 === 0 ? 0xec : 0x11);
	}
	return codewords;
}

function appendBits(bits: number[], value: number, length: number) {
	for (let bitIndex = length - 1; bitIndex >= 0; bitIndex -= 1) {
		bits.push((value >>> bitIndex) & 1);
	}
}

const gfTables = createGaloisTables();

function createGaloisTables() {
	const expTable = new Array<number>(512).fill(0);
	const logTable = new Array<number>(256).fill(0);
	let value = 1;
	for (let exponent = 0; exponent < 255; exponent += 1) {
		expTable[exponent] = value;
		logTable[value] = exponent;
		value <<= 1;
		if (value & 0x100) value ^= 0x11d;
	}
	for (let exponent = 255; exponent < 512; exponent += 1) {
		expTable[exponent] = expTable[exponent - 255];
	}
	return { expTable, logTable };
}

function gfMultiply(left: number, right: number) {
	if (left === 0 || right === 0) return 0;
	return gfTables.expTable[gfTables.logTable[left] + gfTables.logTable[right]];
}

function computeErrorCodewords(dataCodewords: number[], errorCount: number) {
	const generator = buildGeneratorPolynomial(errorCount);
	const message = [...dataCodewords, ...new Array<number>(errorCount).fill(0)];
	for (let dataIndex = 0; dataIndex < dataCodewords.length; dataIndex += 1) {
		const factor = message[dataIndex];
		if (factor === 0) continue;
		for (let generatorIndex = 0; generatorIndex < generator.length; generatorIndex += 1) {
			message[dataIndex + generatorIndex] ^= gfMultiply(generator[generatorIndex], factor);
		}
	}
	return message.slice(dataCodewords.length);
}

function buildGeneratorPolynomial(errorCount: number) {
	let polynomial = [1];
	for (let degree = 0; degree < errorCount; degree += 1) {
		const nextPolynomial = new Array<number>(polynomial.length + 1).fill(0);
		const root = gfTables.expTable[degree];
		for (let index = 0; index < polynomial.length; index += 1) {
			nextPolynomial[index] ^= polynomial[index];
			nextPolynomial[index + 1] ^= gfMultiply(polynomial[index], root);
		}
		polynomial = nextPolynomial;
	}
	return polynomial;
}

function buildModules(spec: QrVersionSpec, codewords: number[]) {
	const modules = Array.from({ length: spec.size }, () => new Array<boolean>(spec.size).fill(false));
	const reserved = Array.from({ length: spec.size }, () => new Array<boolean>(spec.size).fill(false));
	const setModule = (row: number, column: number, dark: boolean, reserve = true) => {
		if (row < 0 || column < 0 || row >= spec.size || column >= spec.size) return;
		modules[row][column] = dark;
		if (reserve) reserved[row][column] = true;
	};

	placeFinderPattern(setModule, 0, 0);
	placeFinderPattern(setModule, 0, spec.size - 7);
	placeFinderPattern(setModule, spec.size - 7, 0);
	if (spec.alignmentCenter !== null) {
		placeAlignmentPattern(setModule, spec.alignmentCenter, spec.alignmentCenter);
	}
	placeTimingPatterns(setModule, spec.size);
	reserveFormatAreas(setModule, spec.size);
	setModule(4 * spec.version + 9, 8, true);

	placeCodewords(modules, reserved, codewords);
	placeFormatBits(setModule, spec.size);
	setModule(4 * spec.version + 9, 8, true);
	return modules;
}

function placeFinderPattern(
	setModule: (row: number, column: number, dark: boolean, reserve?: boolean) => void,
	row: number,
	column: number
) {
	for (let rowOffset = -1; rowOffset <= 7; rowOffset += 1) {
		for (let columnOffset = -1; columnOffset <= 7; columnOffset += 1) {
			const inPattern = rowOffset >= 0 && rowOffset <= 6 && columnOffset >= 0 && columnOffset <= 6;
			const dark =
				inPattern &&
				(rowOffset === 0 ||
					rowOffset === 6 ||
					columnOffset === 0 ||
					columnOffset === 6 ||
					(rowOffset >= 2 && rowOffset <= 4 && columnOffset >= 2 && columnOffset <= 4));
			setModule(row + rowOffset, column + columnOffset, dark);
		}
	}
}

function placeAlignmentPattern(
	setModule: (row: number, column: number, dark: boolean, reserve?: boolean) => void,
	centerRow: number,
	centerColumn: number
) {
	for (let rowOffset = -2; rowOffset <= 2; rowOffset += 1) {
		for (let columnOffset = -2; columnOffset <= 2; columnOffset += 1) {
			const distance = Math.max(Math.abs(rowOffset), Math.abs(columnOffset));
			setModule(centerRow + rowOffset, centerColumn + columnOffset, distance === 2 || distance === 0);
		}
	}
}

function placeTimingPatterns(
	setModule: (row: number, column: number, dark: boolean, reserve?: boolean) => void,
	size: number
) {
	for (let index = 8; index < size - 8; index += 1) {
		setModule(6, index, index % 2 === 0);
		setModule(index, 6, index % 2 === 0);
	}
}

function reserveFormatAreas(
	setModule: (row: number, column: number, dark: boolean, reserve?: boolean) => void,
	size: number
) {
	for (let index = 0; index <= 8; index += 1) {
		if (index !== 6) {
			setModule(8, index, false);
			setModule(index, 8, false);
		}
	}
	for (let index = size - 8; index < size; index += 1) setModule(8, index, false);
	for (let index = size - 7; index < size; index += 1) setModule(index, 8, false);
}

function placeCodewords(modules: boolean[][], reserved: boolean[][], codewords: number[]) {
	const size = modules.length;
	const bits: number[] = [];
	for (const codeword of codewords) appendBits(bits, codeword, 8);

	let bitIndex = 0;
	let upward = true;
	for (let column = size - 1; column > 0; column -= 2) {
		if (column === 6) column -= 1;
		for (let rowOffset = 0; rowOffset < size; rowOffset += 1) {
			const row = upward ? size - 1 - rowOffset : rowOffset;
			for (let columnOffset = 0; columnOffset < 2; columnOffset += 1) {
				const currentColumn = column - columnOffset;
				if (reserved[row][currentColumn]) continue;
				let dark = bitIndex < bits.length ? bits[bitIndex] === 1 : false;
				if ((row + currentColumn) % 2 === 0) dark = !dark;
				modules[row][currentColumn] = dark;
				bitIndex += 1;
			}
		}
		upward = !upward;
	}
}

function placeFormatBits(
	setModule: (row: number, column: number, dark: boolean, reserve?: boolean) => void,
	size: number
) {
	for (let bitIndex = 0; bitIndex < 15; bitIndex += 1) {
		const dark = ((FORMAT_BITS_LEVEL_L_MASK_0 >>> bitIndex) & 1) === 1;
		if (bitIndex < 6) setModule(8, bitIndex, dark);
		else if (bitIndex < 8) setModule(8, bitIndex + 1, dark);
		else setModule(8, size - 15 + bitIndex, dark);

		if (bitIndex < 8) setModule(size - 1 - bitIndex, 8, dark);
		else if (bitIndex < 9) setModule(15 - bitIndex, 8, dark);
		else setModule(14 - bitIndex, 8, dark);
	}
}

function svgDataUrl(modules: boolean[][]) {
	const quietZone = 4;
	const size = modules.length + quietZone * 2;
	const pathParts: string[] = [];
	for (let row = 0; row < modules.length; row += 1) {
		for (let column = 0; column < modules.length; column += 1) {
			if (modules[row][column]) {
				pathParts.push(`M${column + quietZone},${row + quietZone}h1v1h-1z`);
			}
		}
	}
	const svg = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 ${size} ${size}" shape-rendering="crispEdges"><rect width="${size}" height="${size}" fill="#fff"/><path fill="#111827" d="${pathParts.join('')}"/></svg>`;
	return `data:image/svg+xml;charset=utf-8,${encodeURIComponent(svg)}`;
}
