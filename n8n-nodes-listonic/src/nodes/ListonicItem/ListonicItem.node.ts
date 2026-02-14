import {
	IDataObject,
	IExecuteFunctions,
	INodeExecutionData,
	INodeType,
	INodeTypeDescription,
} from 'n8n-workflow';
import { listonicRequest, LISTS_ENDPOINT } from '../../shared/client';

export class ListonicItem implements INodeType {
	description: INodeTypeDescription = {
		displayName: 'Listonic Item',
		name: 'listonicItem',
		icon: 'file:listonic.svg',
		group: ['transform'],
		version: 1,
		subtitle: '={{$parameter["operation"]}}',
		description: 'Manage items in Listonic shopping lists',
		defaults: {
			name: 'Listonic Item',
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
					{ name: 'Get All', value: 'getAll', description: 'Get all items in a list' },
					{ name: 'Add', value: 'add', description: 'Add an item to a list' },
					{ name: 'Update', value: 'update', description: 'Update an item' },
					{ name: 'Delete', value: 'delete', description: 'Delete an item' },
					{
						name: 'Clear Checked',
						value: 'clearChecked',
						description: 'Delete all checked items from a list',
					},
				],
				default: 'getAll',
			},
			{
				displayName: 'List ID',
				name: 'listId',
				type: 'string',
				default: '',
				required: true,
			},
			{
				displayName: 'Item ID',
				name: 'itemId',
				type: 'string',
				default: '',
				required: true,
				displayOptions: {
					show: { operation: ['update', 'delete'] },
				},
			},
			{
				displayName: 'Name',
				name: 'name',
				type: 'string',
				default: '',
				required: true,
				displayOptions: {
					show: { operation: ['add'] },
				},
			},
			{
				displayName: 'Additional Fields',
				name: 'additionalFields',
				type: 'collection',
				placeholder: 'Add Field',
				default: {},
				displayOptions: {
					show: { operation: ['add', 'update'] },
				},
				options: [
					{
						displayName: 'Amount',
						name: 'amount',
						type: 'string',
						default: '',
					},
					{
						displayName: 'Unit',
						name: 'unit',
						type: 'string',
						default: '',
					},
					{
						displayName: 'Description',
						name: 'description',
						type: 'string',
						default: '',
					},
					{
						displayName: 'Checked',
						name: 'checked',
						type: 'boolean',
						default: false,
					},
					{
						displayName: 'Name',
						name: 'name',
						type: 'string',
						default: '',
						description: 'Rename the item (update only)',
					},
				],
			},
		],
	};

	async execute(this: IExecuteFunctions): Promise<INodeExecutionData[][]> {
		const items = this.getInputData();
		const returnData: INodeExecutionData[] = [];
		const operation = this.getNodeParameter('operation', 0) as string;

		for (let i = 0; i < items.length; i++) {
			const listId = (this.getNodeParameter('listId', i) as string).trim();
			const basePath = `${LISTS_ENDPOINT}/${listId}/items`;
			let result: unknown;

			switch (operation) {
				case 'getAll':
					result = await listonicRequest(this, 'GET', basePath);
					if (Array.isArray(result)) {
						for (const item of result) {
							returnData.push({ json: item as IDataObject });
						}
						continue;
					}
					break;

				case 'add': {
					const name = (this.getNodeParameter('name', i) as string).trim();
					const extra = this.getNodeParameter('additionalFields', i, {}) as {
						amount?: string;
						unit?: string;
						description?: string;
					};
					const payload: Record<string, unknown> = { Name: name };
					if (extra.amount) payload.Amount = String(extra.amount).trim();
					if (extra.unit) payload.Unit = String(extra.unit).trim();
					if (extra.description) payload.Description = String(extra.description).trim();
					result = await listonicRequest(this, 'POST', basePath, payload);
					break;
				}

				case 'update': {
					const itemId = (this.getNodeParameter('itemId', i) as string).trim();
					const extra = this.getNodeParameter('additionalFields', i, {}) as {
						amount?: string;
						unit?: string;
						description?: string;
						checked?: boolean;
						name?: string;
					};
					const payload: Record<string, unknown> = {};
					if (extra.name) payload.Name = String(extra.name).trim();
					if (extra.amount) payload.Amount = String(extra.amount).trim();
					if (extra.unit) payload.Unit = String(extra.unit).trim();
					if (extra.description) payload.Description = String(extra.description).trim();
					if (extra.checked !== undefined) payload.Checked = extra.checked ? 1 : 0;
					result = await listonicRequest(this, 'PATCH', `${basePath}/${itemId}`, payload);
					break;
				}

				case 'delete': {
					const itemId = (this.getNodeParameter('itemId', i) as string).trim();
					await listonicRequest(this, 'DELETE', `${basePath}/${itemId}`);
					result = { success: true };
					break;
				}

				case 'clearChecked': {
					const allItems = (await listonicRequest(this, 'GET', basePath)) as Array<{
						Id: string;
						Checked: number;
					}>;
					const deleted: string[] = [];
					for (const item of allItems) {
						if (item.Checked === 1) {
							await listonicRequest(this, 'DELETE', `${basePath}/${item.Id}`);
							deleted.push(item.Id);
						}
					}
					result = { deleted, count: deleted.length };
					break;
				}
			}

			returnData.push({ json: (result ?? { success: true }) as IDataObject });
		}

		return [returnData];
	}
}
