import {
	IExecuteFunctions,
	INodeExecutionData,
	INodeType,
	INodeTypeDescription,
} from 'n8n-workflow';
import { parseRecipe } from '../../shared/parser';

export class CooklangParse implements INodeType {
	description: INodeTypeDescription = {
		displayName: 'Cooklang Parse',
		name: 'cooklangParse',
		icon: 'file:cooklang.svg',
		group: ['transform'],
		version: 1,
		subtitle: 'Parse Cooklang recipe',
		description: 'Parse raw .cook text into structured recipe data',
		defaults: {
			name: 'Cooklang Parse',
		},
		inputs: ['main'],
		outputs: ['main'],
		properties: [
			{
				displayName: 'Recipe Text',
				name: 'recipeText',
				type: 'string',
				typeOptions: {
					rows: 10,
				},
				default: '',
				required: true,
				description: 'Raw Cooklang recipe text to parse',
			},
		],
	};

	async execute(this: IExecuteFunctions): Promise<INodeExecutionData[][]> {
		const items = this.getInputData();
		const returnData: INodeExecutionData[] = [];

		for (let i = 0; i < items.length; i++) {
			const recipeText = this.getNodeParameter('recipeText', i) as string;
			const parsed = parseRecipe(recipeText);
			returnData.push({ json: { ...items[i].json, ...parsed } });
		}

		return [returnData];
	}
}
