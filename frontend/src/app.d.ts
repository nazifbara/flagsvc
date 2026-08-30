// See https://svelte.dev/docs/kit/types#app.d.ts
/// <reference types="svelte-clerk/env" />

import type { Doc } from '$convex/_generated/dataModel';
import type { ConvexHttpClient } from 'convex/browser';
import type Stripe from 'stripe';

// for information about these interfaces
declare global {
	namespace App {
		// interface Error {}
		interface Locals {
			convex: ConvexHttpClient;
			stripe: Stripe;
			authToken?: string;
			userId?: string;
			profile?: Doc<'profiles'>;
		}
		// interface PageData {}
		// interface PageState {}
		// interface Platform {}
	}
}

export {};
