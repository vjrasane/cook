import { IExecuteFunctions, IHttpRequestMethods, IHttpRequestOptions } from 'n8n-workflow';

const API_BASE = 'https://api.listonic.com';
const LOGIN_ENDPOINT = '/api/loginextended';
const LISTS_ENDPOINT = '/api/lists';

const CLIENT_ID = 'listonicv2';
const CLIENT_SECRET = 'fjdfsoj9874jdfhjkh34jkhffdfff';
const REDIRECT_URI = 'https://listonicv2api.jestemkucharzem.pl';

const CLIENT_AUTH = Buffer.from(`${CLIENT_ID}:${CLIENT_SECRET}`).toString('base64');

interface TokenData {
	access_token: string;
	refresh_token: string;
	expires_in: number;
}

let cachedToken: { accessToken: string; refreshToken: string; expiresAt: number } | null = null;

async function authenticate(context: IExecuteFunctions): Promise<string> {
	if (cachedToken && Date.now() < cachedToken.expiresAt) {
		return cachedToken.accessToken;
	}

	if (cachedToken?.refreshToken) {
		try {
			return await refreshToken(context, cachedToken.refreshToken);
		} catch {
			cachedToken = null;
		}
	}

	return login(context);
}

function encodeForm(params: Record<string, string>): string {
	return Object.entries(params)
		.map(([k, v]) => `${encodeURIComponent(k)}=${encodeURIComponent(v)}`)
		.join('&');
}

async function login(context: IExecuteFunctions): Promise<string> {
	const credentials = await context.getCredentials('listonicCredentials');

	const body = encodeForm({
		username: credentials.email as string,
		password: credentials.password as string,
		client_id: CLIENT_ID,
		client_secret: CLIENT_SECRET,
		redirect_uri: REDIRECT_URI,
	});

	const response = await context.helpers.httpRequest({
		method: 'POST',
		url: `${API_BASE}${LOGIN_ENDPOINT}?provider=password&autoMerge=1&autoDestruct=1`,
		headers: {
			'Content-Type': 'application/x-www-form-urlencoded',
			Accept: 'application/json',
			clientauthorization: `Bearer ${CLIENT_AUTH}`,
		},
		body,
	});

	const data = response as TokenData;
	cachedToken = {
		accessToken: data.access_token,
		refreshToken: data.refresh_token,
		expiresAt: Date.now() + data.expires_in * 1000,
	};
	return data.access_token;
}

async function refreshToken(context: IExecuteFunctions, refresh: string): Promise<string> {
	const body = encodeForm({
		grant_type: 'refresh_token',
		refresh_token: refresh,
		client_id: CLIENT_ID,
		client_secret: CLIENT_SECRET,
	});

	const response = await context.helpers.httpRequest({
		method: 'POST',
		url: `${API_BASE}${LOGIN_ENDPOINT}`,
		headers: {
			'Content-Type': 'application/x-www-form-urlencoded',
			Accept: 'application/json',
			clientauthorization: `Bearer ${CLIENT_AUTH}`,
		},
		body,
	});

	const data = response as TokenData;
	cachedToken = {
		accessToken: data.access_token,
		refreshToken: data.refresh_token,
		expiresAt: Date.now() + data.expires_in * 1000,
	};
	return data.access_token;
}

export async function listonicRequest(
	context: IExecuteFunctions,
	method: IHttpRequestMethods,
	path: string,
	body?: object,
): Promise<unknown> {
	const token = await authenticate(context);

	const options: IHttpRequestOptions = {
		method,
		url: `${API_BASE}${path}`,
		headers: {
			'Content-Type': 'application/json',
			Accept: 'application/json',
			Authorization: `Bearer ${token}`,
		},
		json: true,
	};

	if (body) {
		options.body = JSON.stringify(body);
	}

	return context.helpers.httpRequest(options);
}

export { LISTS_ENDPOINT };
