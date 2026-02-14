import { ICredentialType, INodeProperties } from 'n8n-workflow';

export class ListonicCredentials implements ICredentialType {
	name = 'listonicCredentials';
	displayName = 'Listonic Credentials';

	properties: INodeProperties[] = [
		{
			displayName: 'Email',
			name: 'email',
			type: 'string',
			default: '',
			required: true,
		},
		{
			displayName: 'Password',
			name: 'password',
			type: 'string',
			typeOptions: { password: true },
			default: '',
			required: true,
		},
	];
}
