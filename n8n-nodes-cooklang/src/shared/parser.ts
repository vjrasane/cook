import matter from "gray-matter";
import { Recipe } from "@cooklang/cooklang-ts";

export function parseRecipe(source: string) {
  const { content, data } = matter(source);
  const recipe = new Recipe(content);
  return {
    metadata: { ...data, ...recipe.metadata },
    ingredients: recipe.ingredients,
    steps: recipe.steps,
    cookwares: recipe.cookwares,
    shoppingList: recipe.shoppingList,
  };
}
