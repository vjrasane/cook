package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"
)

var Version = "dev"

func requireEnv() *Client {
	email := os.Getenv("LISTONIC_EMAIL")
	password := os.Getenv("LISTONIC_PASSWORD")
	if email == "" || password == "" {
		printError(fmt.Errorf("LISTONIC_EMAIL and LISTONIC_PASSWORD must be set"))
	}

	client := NewClient(email, password)
	if err := client.Authenticate(); err != nil {
		printError(err)
	}
	return client
}

func main() {
	rootCmd := &cobra.Command{
		Use:   "listonic",
		Short: "Listonic shopping list CLI",
	}

	rootCmd.AddCommand(
		listsCmd(),
		itemsCmd(),
		addCmd(),
		checkCmd(),
		uncheckCmd(),
		deleteItemCmd(),
		versionCmd(),
	)

	rootCmd.Execute()
}

func listsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "lists",
		Short: "Manage shopping lists",
	}

	cmd.AddCommand(
		listsGetCmd(),
		listsCreateCmd(),
		listsDeleteCmd(),
		listsUpdateCmd(),
		listsClearCmd(),
	)

	return cmd
}

func listsGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get [list]",
		Short: "Get shopping lists or a single list",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			client := requireEnv()
			if len(args) == 1 {
				listID, err := client.ResolveListID(args[0])
				if err != nil {
					printError(err)
				}
				l, err := client.GetList(listID)
				if err != nil {
					printError(err)
				}
				printSuccess(l)
				return
			}
			lists, err := client.GetLists()
			if err != nil {
				printError(err)
			}
			printSuccess(lists)
		},
	}
}

func listsCreateCmd() *cobra.Command {
	var name string
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new shopping list",
		Run: func(cmd *cobra.Command, args []string) {
			client := requireEnv()
			list, err := client.CreateList(name)
			if err != nil {
				printError(err)
			}
			printSuccess(list)
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "List name")
	cmd.MarkFlagRequired("name")
	return cmd
}

func listsDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <list>",
		Short: "Delete a shopping list",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			client := requireEnv()
			listID, err := client.ResolveListID(args[0])
			if err != nil {
				printError(err)
			}
			if err := client.DeleteList(listID); err != nil {
				printError(err)
			}
			printSuccess(map[string]string{"deleted": listID})
		},
	}
}

func listsUpdateCmd() *cobra.Command {
	var name string
	cmd := &cobra.Command{
		Use:   "update <list>",
		Short: "Update a shopping list",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			client := requireEnv()
			listID, err := client.ResolveListID(args[0])
			if err != nil {
				printError(err)
			}
			if err := client.UpdateList(listID, name); err != nil {
				printError(err)
			}
			printSuccess(map[string]string{"updated": listID, "name": name})
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "New list name")
	cmd.MarkFlagRequired("name")
	return cmd
}

func listsClearCmd() *cobra.Command {
	var all, checked bool
	cmd := &cobra.Command{
		Use:   "clear <list>",
		Short: "Remove items from a list",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			client := requireEnv()
			listID, err := client.ResolveListID(args[0])
			if err != nil {
				printError(err)
			}
			deleted, err := client.ClearItems(listID, checked)
			if err != nil {
				printError(err)
			}
			printSuccess(map[string]any{"deleted": len(deleted), "itemIds": deleted})
		},
	}
	cmd.Flags().BoolVar(&all, "all", false, "Remove all items")
	cmd.Flags().BoolVar(&checked, "checked", false, "Remove only checked items")
	cmd.MarkFlagsOneRequired("all", "checked")
	cmd.MarkFlagsMutuallyExclusive("all", "checked")
	return cmd
}

func itemsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "items <list>",
		Short: "List items in a list",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			client := requireEnv()
			listID, err := client.ResolveListID(args[0])
			if err != nil {
				printError(err)
			}
			items, err := client.GetListItems(listID)
			if err != nil {
				printError(err)
			}
			printSuccess(items)
		},
	}
}

func addCmd() *cobra.Command {
	var list, name, amount, unit string
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add an item to a list",
		Run: func(cmd *cobra.Command, args []string) {
			client := requireEnv()
			listID, err := client.ResolveListID(list)
			if err != nil {
				printError(err)
			}
			item, err := client.AddItem(listID, AddItemRequest{
				Name:   name,
				Amount: amount,
				Unit:   unit,
			})
			if err != nil {
				printError(err)
			}
			printSuccess(item)
		},
	}
	cmd.Flags().StringVar(&list, "list", "", "List name or ID")
	cmd.Flags().StringVar(&name, "name", "", "Item name")
	cmd.Flags().StringVar(&amount, "amount", "", "Quantity")
	cmd.Flags().StringVar(&unit, "unit", "", "Unit of measurement")
	cmd.MarkFlagRequired("list")
	cmd.MarkFlagRequired("name")
	return cmd
}

func checkCmd() *cobra.Command {
	var list, item string
	cmd := &cobra.Command{
		Use:   "check",
		Short: "Check off an item",
		Run: func(cmd *cobra.Command, args []string) {
			client := requireEnv()
			listID, err := client.ResolveListID(list)
			if err != nil {
				printError(err)
			}
			checked := 1
			if err := client.UpdateItem(listID, item, UpdateItemRequest{Checked: &checked}); err != nil {
				printError(err)
			}
			itemID, _ := strconv.ParseInt(item, 10, 64)
			printSuccess(Item{Id: item, IdAsNumber: itemID, Checked: 1})
		},
	}
	cmd.Flags().StringVar(&list, "list", "", "List name or ID")
	cmd.Flags().StringVar(&item, "item", "", "Item ID")
	cmd.MarkFlagRequired("list")
	cmd.MarkFlagRequired("item")
	return cmd
}

func uncheckCmd() *cobra.Command {
	var list, item string
	cmd := &cobra.Command{
		Use:   "uncheck",
		Short: "Uncheck an item",
		Run: func(cmd *cobra.Command, args []string) {
			client := requireEnv()
			listID, err := client.ResolveListID(list)
			if err != nil {
				printError(err)
			}
			unchecked := 0
			if err := client.UpdateItem(listID, item, UpdateItemRequest{Checked: &unchecked}); err != nil {
				printError(err)
			}
			itemID, _ := strconv.ParseInt(item, 10, 64)
			printSuccess(Item{Id: item, IdAsNumber: itemID, Checked: 0})
		},
	}
	cmd.Flags().StringVar(&list, "list", "", "List name or ID")
	cmd.Flags().StringVar(&item, "item", "", "Item ID")
	cmd.MarkFlagRequired("list")
	cmd.MarkFlagRequired("item")
	return cmd
}

func deleteItemCmd() *cobra.Command {
	var list, item string
	cmd := &cobra.Command{
		Use:   "delete-item",
		Short: "Remove an item from a list",
		Run: func(cmd *cobra.Command, args []string) {
			client := requireEnv()
			listID, err := client.ResolveListID(list)
			if err != nil {
				printError(err)
			}
			if err := client.DeleteItem(listID, item); err != nil {
				printError(err)
			}
			printSuccess(map[string]string{"deleted": item})
		},
	}
	cmd.Flags().StringVar(&list, "list", "", "List name or ID")
	cmd.Flags().StringVar(&item, "item", "", "Item ID")
	cmd.MarkFlagRequired("list")
	cmd.MarkFlagRequired("item")
	return cmd
}

func versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(Version)
		},
	}
}
