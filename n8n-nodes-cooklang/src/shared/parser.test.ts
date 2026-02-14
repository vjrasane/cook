import { describe, it, expect } from 'vitest';
import { readFileSync } from 'fs';
import { resolve } from 'path';
import { parseRecipe } from './parser';

describe('parseRecipe', () => {
	it('extracts scalar frontmatter values into metadata', () => {
		const source = `---
servings: 4
course: main
time: 35 minutes
---

Dice @potato{2%pcs}.`;

		const result = parseRecipe(source);
		expect(result.metadata.servings).toBe(4);
		expect(result.metadata.course).toBe('main');
		expect(result.metadata.time).toBe('35 minutes');
	});

	it('extracts list frontmatter values as string arrays', () => {
		const source = `---
tags:
  - curry
  - vegan
---

Dice @potato{2%pcs}.`;

		const result = parseRecipe(source);
		expect(result.metadata.tags).toEqual(['curry', 'vegan']);
	});

	it('parses plain cooklang without frontmatter', () => {
		const source = `Dice @potato{2%pcs} and fry in @oil{}.`;

		const result = parseRecipe(source);
		expect(result.ingredients).toHaveLength(2);
		expect(result.ingredients[0].name).toBe('potato');
		expect(result.steps).toHaveLength(1);
	});

	it('gives cooklang-native metadata precedence over frontmatter', () => {
		const source = `---
servings: 4
---

>> servings: 6

Dice @potato{2%pcs}.`;

		const result = parseRecipe(source);
		expect(result.metadata.servings).toBe('6');

	});

	it('preserves ingredients and steps when frontmatter is present', () => {
		const source = `---
servings: 4
tags:
  - test
---

Dice @potato{2%pcs} and @onion{1}. Cook in #pot{} for ~{20%minutes}.`;

		const result = parseRecipe(source);
		expect(result.ingredients).toHaveLength(2);
		expect(result.ingredients.map((i) => i.name)).toEqual(['potato', 'onion']);
		expect(result.cookwares).toHaveLength(1);
		expect(result.cookwares[0].name).toBe('pot');
		expect(result.steps).toHaveLength(1);
	});

	it('parses a real recipe file end-to-end', () => {
		const recipePath = resolve(__dirname, '../../../recipes/Bataatticurry.cook');
		const source = readFileSync(recipePath, 'utf-8');
		const result = parseRecipe(source);

		expect(result.metadata.servings).toBe(4);
		expect(result.metadata.time).toBe('35 minutes');
		expect(result.metadata.course).toBe('main');
		expect(result.metadata.tags).toEqual(['curry', 'vegan']);

		expect(result.ingredients.length).toBeGreaterThan(0);
		const names = result.ingredients.map((i) => i.name);
		expect(names).toContain('bataatti');
		expect(names).toContain('kookosmaitoa');
		expect(names).toContain('pinaatti');

		expect(result.steps.length).toBeGreaterThan(0);
	});
});
