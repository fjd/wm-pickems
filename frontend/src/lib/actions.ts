/** Toggle a `scrolled` class on the node once the window is scrolled past
 *  `threshold` px. Used to collapse a sticky page header (hide the big title
 *  and intro, keep the kicker line + tabs) while scrolling. */
export function collapseOnScroll(node: HTMLElement, threshold = 16) {
	const update = () =>
		node.classList.toggle('scrolled', window.scrollY > threshold);
	update();
	window.addEventListener('scroll', update, { passive: true });
	return {
		destroy() {
			window.removeEventListener('scroll', update);
		}
	};
}
