import { House, Volleyball, Telescope, Network, Trophy } from '@lucide/svelte';
import type { Component } from 'svelte';

export interface NavItem {
	href: string;
	key: string;
	icon: Component;
}

export const navItems: NavItem[] = [
	{ href: '/', key: 'home', icon: House },
	{ href: '/tips', key: 'tips', icon: Volleyball },
	{ href: '/forecast', key: 'forecast', icon: Telescope },
	{ href: '/tournament', key: 'bracket', icon: Network },
	{ href: '/leagues', key: 'leagues', icon: Trophy }
];

export function isActive(href: string, path: string): boolean {
	return href === '/' ? path === '/' : path.startsWith(href);
}
