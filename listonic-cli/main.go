package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"
)

var Version = "dev"

func authenticatedClient() *Client {
	client := NewClient(os.Getenv("LISTONIC_EMAIL"), os.Getenv("LISTONIC_PASSWORD"))
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
		listCmd(),
		itemCmd(),
		versionCmd(),
	)

	rootCmd.Execute()
}

func listCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "Manage shopping lists",
	}

	cmd.AddCommand(
		listGetCmd(),
		listCreateCmd(),
		listDeleteCmd(),
		listUpdateCmd(),
		listClearCmd(),
		listItemsCmd(),
	)

	return cmd
}

func listGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get [list]",
		Short: "Get shopping lists or a single list",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			client := authenticatedClient()
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

func listCreateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "create <name>",
		Short: "Create a new shopping list",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			client := authenticatedClient()
			list, err := client.CreateList(args[0])
			if err != nil {
				printError(err)
			}
			printSuccess(list)
		},
	}
}

func listDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <list>",
		Short: "Delete a shopping list",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			client := authenticatedClient()
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

func listUpdateCmd() *cobra.Command {
	var name string
	cmd := &cobra.Command{
		Use:   "update <list>",
		Short: "Update a shopping list",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			client := authenticatedClient()
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

func listClearCmd() *cobra.Command {
	var all, checked bool
	cmd := &cobra.Command{
		Use:   "clear <list>",
		Short: "Remove items from a list",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			client := authenticatedClient()
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

func listItemsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "items <list>",
		Short: "List items in a list",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			client := authenticatedClient()
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

func itemCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "item",
		Short: "Manage shopping list items",
	}

	cmd.AddCommand(
		itemAddCmd(),
		itemUpdateCmd(),
		itemCheckCmd(),
		itemDeleteCmd(),
	)

	return cmd
}

func itemAddCmd() *cobra.Command {
	var list, amount, unit string
	cmd := &cobra.Command{
		Use:   "add <name>",
		Short: "Add an item to a list",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			client := authenticatedClient()
			listID, err := client.ResolveListID(list)
			if err != nil {
				printError(err)
			}
			item, err := client.AddItem(listID, AddItemRequest{
				Name:   args[0],
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
	cmd.Flags().StringVar(&amount, "amount", "", "Quantity")
	cmd.Flags().StringVar(&unit, "unit", "", "Unit of measurement")
	cmd.MarkFlagRequired("list")
	return cmd
}

func itemUpdateCmd() *cobra.Command {
	var list string
	var check, uncheck bool
	cmd := &cobra.Command{
		Use:   "update <item>",
		Short: "Update an item",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			client := authenticatedClient()
			listID, err := client.ResolveListID(list)
			if err != nil {
				printError(err)
			}
			var update UpdateItemRequest
			if check {
				v := 1
				update.Checked = &v
			} else if uncheck {
				v := 0
				update.Checked = &v
			}
			if err := client.UpdateItem(listID, args[0], update); err != nil {
				printError(err)
			}
			itemID, _ := strconv.ParseInt(args[0], 10, 64)
			checked := 0
			if check {
				checked = 1
			}
			printSuccess(Item{Id: args[0], IdAsNumber: itemID, Checked: checked})
		},
	}
	cmd.Flags().StringVar(&list, "list", "", "List name or ID")
	cmd.Flags().BoolVar(&check, "check", false, "Mark as checked")
	cmd.Flags().BoolVar(&uncheck, "uncheck", false, "Mark as unchecked")
	cmd.MarkFlagRequired("list")
	cmd.MarkFlagsMutuallyExclusive("check", "uncheck")
	return cmd
}

func itemCheckCmd() *cobra.Command {
	var list string
	cmd := &cobra.Command{
		Use:   "check <item>",
		Short: "Check off an item",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			client := authenticatedClient()
			listID, err := client.ResolveListID(list)
			if err != nil {
				printError(err)
			}
			checked := 1
			if err := client.UpdateItem(listID, args[0], UpdateItemRequest{Checked: &checked}); err != nil {
				printError(err)
			}
			itemID, _ := strconv.ParseInt(args[0], 10, 64)
			printSuccess(Item{Id: args[0], IdAsNumber: itemID, Checked: 1})
		},
	}
	cmd.Flags().StringVar(&list, "list", "", "List name or ID")
	cmd.MarkFlagRequired("list")
	return cmd
}

func itemDeleteCmd() *cobra.Command {
	var list string
	cmd := &cobra.Command{
		Use:   "delete <item>",
		Short: "Remove an item from a list",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			client := authenticatedClient()
			listID, err := client.ResolveListID(list)
			if err != nil {
				printError(err)
			}
			if err := client.DeleteItem(listID, args[0]); err != nil {
				printError(err)
			}
			printSuccess(map[string]string{"deleted": args[0]})
		},
	}
	cmd.Flags().StringVar(&list, "list", "", "List name or ID")
	cmd.MarkFlagRequired("list")
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
