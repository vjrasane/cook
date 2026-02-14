import {
	IDataObject,
	IExecuteFunctions,
	INodeExecutionData,
	INodeType,
	INodeTypeDescription,
} from 'n8n-workflow';
import { listonicRequest, LISTS_ENDPOINT } from '../../shared/client';

export class ListonicList implements INodeType {
	description: INodeTypeDescription = {
		displayName: 'Listonic List',
		name: 'listonicList',
		icon: 'file:listonic.svg',
		group: ['transform'],
		version: 1,
		subtitle: '={{$parameter["operation"]}}',
		description: 'Manage Listonic shopping lists',
		defaults: {
			name: 'Listonic List',
		},
		inputs: ['main'],
		outputs: ['main'],
		credentials: [
			{
				name: 'listonicCredentials',
				required: true,
			},
		],
		properties: [
			{
				displayName: 'Operation',
				name: 'operation',
				type: 'options',
				noDataExpression: true,
				options: [
					{ name: 'Get All', value: 'getAll', description: 'Get all lists' },
					{ name: 'Get One', value: 'get', description: 'Get a single list' },
					{ name: 'Create', value: 'create', description: 'Create a new list' },
					{ name: 'Update', value: 'update', description: 'Rename a list' },
					{ name: 'Delete', value: 'delete', description: 'Delete a list' },
				],
				default: 'getAll',
			},
			{
				displayName: 'List ID',
				name: 'listId',
				type: 'string',
				default: '',
				required: true,
				displayOptions: {
					show: { operation: ['get', 'update', 'delete'] },
				},
			},
			{
				displayName: 'Name',
				name: 'name',
				type: 'string',
				default: '',
				required: true,
				displayOptions: {
					show: { operation: ['create', 'update'] },
				},
			},
		],
	};

	async execute(this: IExecuteFunctions): Promise<INodeExecutionData[][]> {
		const items = this.getInputData();
		const returnData: INodeExecutionData[] = [];
		const operation = this.getNodeParameter('operation', 0) as string;

		for (let i = 0; i < items.length; i++) {
			let result: unknown;

			switch (operation) {
				case 'getAll':
					result = await listonicRequest(
						this,
						'GET',
						`${LISTS_ENDPOINT}?includeShares=true&archive=false&includeItems=false`,
					);
					if (Array.isArray(result)) {
						for (const item of result) {
							returnData.push({ json: item as IDataObject });
						}
						continue;
					}
					break;

				case 'get': {
					const listId = (this.getNodeParameter('listId', i) as string).trim();
					result = await listonicRequest(this, 'GET', `${LISTS_ENDPOINT}/${listId}`);
					break;
				}

				case 'create': {
					const name = (this.getNodeParameter('name', i) as string).trim();
					result = await listonicRequest(this, 'POST', LISTS_ENDPOINT, { Name: name });
					break;
				}

				case 'update': {
					const listId = (this.getNodeParameter('listId', i) as string).trim();
					const name = (this.getNodeParameter('name', i) as string).trim();
					result = await listonicRequest(this, 'PATCH', `${LISTS_ENDPOINT}/${listId}`, {
						Name: name,
					});
					break;
				}

				case 'delete': {
					const listId = (this.getNodeParameter('listId', i) as string).trim();
					await listonicRequest(this, 'DELETE', `${LISTS_ENDPOINT}/${listId}`);
					result = { success: true };
					break;
				}
			}

			returnData.push({ json: (result ?? { success: true }) as IDataObject });
		}

		return [returnData];
	}
}
