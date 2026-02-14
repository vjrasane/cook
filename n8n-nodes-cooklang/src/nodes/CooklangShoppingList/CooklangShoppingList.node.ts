import {
	IExecuteFunctions,
	INodeExecutionData,
	INodeType,
	INodeTypeDescription,
} from 'n8n-workflow';

interface Ingredient {
	name: string;
	quantity?: string | number;
	units?: string;
}

interface AisleMap {
	[ingredient: string]: string;
}

export class CooklangShoppingList implements INodeType {
	description: INodeTypeDescription = {
		displayName: 'Cooklang Shopping List',
		name: 'cooklangShoppingList',
		icon: 'file:cooklang.svg',
		group: ['transform'],
		version: 1,
		subtitle: 'Generate shopping list from recipes',
		description:
			'Combine ingredients from parsed Cooklang recipes into a deduplicated shopping list',
		defaults: {
			name: 'Cooklang Shopping List',
		},
		inputs: ['main'],
		outputs: ['main'],
		properties: [
			{
				displayName: 'Aisle Config',
				name: 'aisleConfig',
				type: 'string',
				typeOptions: {
					rows: 10,
				},
				default: '',
				description:
					'INI-format aisle configuration for categorizing ingredients into store sections. Leave empty to skip categorization.',
			},
		],
	};

	async execute(this: IExecuteFunctions): Promise<INodeExecutionData[][]> {
		const items = this.getInputData();
		const aisleConfig = this.getNodeParameter('aisleConfig', 0, '') as string;
		const aisleMap = parseAisleConfig(aisleConfig);
		const merged = new Map<string, { quantity: string; units: string; aisle: string }>();

		for (const item of items) {
			const ingredients = (item.json.ingredients ?? []) as Ingredient[];
			for (const ing of ingredients) {
				const key = ing.name.toLowerCase();
				const existing = merged.get(key);
				const qty = ing.quantity != null ? String(ing.quantity) : '';
				const units = ing.units ?? '';

				if (existing) {
					existing.quantity = mergeQuantities(existing.quantity, qty);
					if (!existing.units && units) existing.units = units;
				} else {
					merged.set(key, {
						quantity: qty,
						units,
						aisle: aisleMap[key] ?? '',
					});
				}
			}
		}

		const returnData: INodeExecutionData[] = [];
		for (const [name, info] of merged) {
			returnData.push({
				json: {
					name,
					quantity: info.quantity,
					units: info.units,
					aisle: info.aisle,
				},
			});
		}

		return [returnData];
	}
}

function mergeQuantities(a: string, b: string): string {
	if (!a) return b;
	if (!b) return a;
	const numA = parseFloat(a);
	const numB = parseFloat(b);
	if (!isNaN(numA) && !isNaN(numB)) return String(numA + numB);
	return `${a}, ${b}`;
}

function parseAisleConfig(config: string): AisleMap {
	const map: AisleMap = {};
	if (!config.trim()) return map;

	let currentAisle = '';
	for (const line of config.split('\n')) {
		const trimmed = line.trim();
		if (!trimmed) continue;
		const sectionMatch = trimmed.match(/^\[(.+)]$/);
		if (sectionMatch) {
			currentAisle = sectionMatch[1];
		} else if (currentAisle) {
			map[trimmed.toLowerCase()] = currentAisle;
		}
	}
	return map;
}
