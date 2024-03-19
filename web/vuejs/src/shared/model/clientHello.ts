export interface ClientHello {
	type: 'hello';
	auth: ClientHelloAuth;
}

interface ClientHelloAuth {
	keycloak: string;
}
