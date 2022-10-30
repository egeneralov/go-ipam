package main

import (
  "context"
  "fmt"
  "net/http"
  "os"
  
  "github.com/bufbuild/connect-go"
  compress "github.com/klauspost/connect-compress"
  "github.com/metal-stack/v"
  "github.com/urfave/cli/v2"
  
  v1 "github.com/metal-stack/go-ipam/api/v1"
  "github.com/metal-stack/go-ipam/api/v1/apiv1connect"
)

func main() {
  
  app := &cli.App{
    Name:    "cli",
    Usage:   "cli for go-ipam",
    Version: v.V.String(),
    Flags: []cli.Flag{
      &cli.StringFlag{
        Name:    "grpc-server-endpoint",
        Value:   "http://localhost:9090",
        Usage:   "gRPC server endpoint",
        EnvVars: []string{"GOIPAM_GRPC_SERVER_ENDPOINT"},
      },
    },
    Commands: []*cli.Command{
      {
        Name:    "prefix",
        Aliases: []string{"p"},
        Usage:   "prefix manipulation",
        Subcommands: []*cli.Command{
          {
            Name:  "create",
            Usage: "create a prefix",
            Flags: []cli.Flag{
              &cli.StringFlag{
                Name: "cidr",
              },
            },
            Action: func(ctx *cli.Context) error {
              c := client(ctx)
              result, err := c.CreatePrefix(context.Background(), connect.NewRequest(&v1.CreatePrefixRequest{
                Cidr: ctx.String("cidr"),
              }))
              
              if err != nil {
                return err
              }
              fmt.Println(result.Msg.Prefix.Cidr)
              return nil
            },
          },
          {
            Name:  "acquire",
            Usage: "acquire a child prefix",
            Flags: []cli.Flag{
              &cli.StringFlag{
                Name: "parent",
              },
              &cli.UintFlag{
                Name: "length",
              },
            },
            Action: func(ctx *cli.Context) error {
              c := client(ctx)
              result, err := c.AcquireChildPrefix(context.Background(), connect.NewRequest(&v1.AcquireChildPrefixRequest{
                Cidr:   ctx.String("parent"),
                Length: uint32(ctx.Uint("length")),
              }))
              
              if err != nil {
                return err
              }
              fmt.Println(result.Msg.Prefix.Cidr)
              return nil
            },
          },
          {
            Name:  "release",
            Usage: "release a child prefix",
            Flags: []cli.Flag{
              &cli.StringFlag{
                Name: "cidr",
              },
            },
            Action: func(ctx *cli.Context) error {
              c := client(ctx)
              result, err := c.ReleaseChildPrefix(context.Background(), connect.NewRequest(&v1.ReleaseChildPrefixRequest{
                Cidr: ctx.String("cidr"),
              }))
              
              if err != nil {
                return err
              }
              if result.Msg == nil || result.Msg.Prefix == nil {
                return fmt.Errorf("result contains no prefix")
              }
              fmt.Println(result.Msg.Prefix.Cidr)
              return nil
            },
          },
          {
            Name:  "list",
            Usage: "list all prefixes",
            Action: func(ctx *cli.Context) error {
              c := client(ctx)
              result, err := c.ListPrefixes(context.Background(), connect.NewRequest(&v1.ListPrefixesRequest{}))
              
              if err != nil {
                return err
              }
              for _, p := range result.Msg.Prefixes {
                fmt.Printf("Prefix:%q parent:%q\n", p.Cidr, p.ParentCidr)
              }
              return nil
            },
          },
          {
            Name:  "delete",
            Usage: "delete a prefix",
            Flags: []cli.Flag{
              &cli.StringFlag{
                Name: "cidr",
              },
            },
            Action: func(ctx *cli.Context) error {
              c := client(ctx)
              result, err := c.DeletePrefix(context.Background(), connect.NewRequest(&v1.DeletePrefixRequest{
                Cidr: ctx.String("cidr"),
              }))
              
              if err != nil {
                return err
              }
              fmt.Println(result.Msg.Prefix.Cidr)
              return nil
            },
          },
        },
      },
      {
        Name:    "ip",
        Aliases: []string{"i"},
        Usage:   "ip manipulation",
        Subcommands: []*cli.Command{
          {
            Name:  "acquire",
            Usage: "acquire a ip",
            Flags: []cli.Flag{
              &cli.StringFlag{
                Name: "prefix",
              },
            },
            Action: func(ctx *cli.Context) error {
              c := client(ctx)
              result, err := c.AcquireIP(context.Background(), connect.NewRequest(&v1.AcquireIPRequest{
                PrefixCidr: ctx.String("prefix"),
              }))
              
              if err != nil {
                return err
              }
              fmt.Println(result.Msg.Ip.Ip)
              return nil
            },
          },
          {
            Name:  "release",
            Usage: "release a ip",
            Flags: []cli.Flag{
              &cli.StringFlag{
                Name: "ip",
              },
              &cli.StringFlag{
                Name: "prefix",
              },
            },
            Action: func(ctx *cli.Context) error {
              c := client(ctx)
              result, err := c.ReleaseIP(context.Background(), connect.NewRequest(&v1.ReleaseIPRequest{
                Ip:         ctx.String("ip"),
                PrefixCidr: ctx.String("prefix"),
              }))
              
              if err != nil {
                return err
              }
              fmt.Println(result.Msg.Ip.Ip)
              return nil
            },
          },
        },
      },
      {
        Name:  "backup",
        Usage: "create and restore a backup",
        Subcommands: []*cli.Command{
          {
            Name:  "create",
            Usage: "create a json file of the whole ipam db for backup purpose",
            Flags: []cli.Flag{
              &cli.StringFlag{
                Name: "file",
              },
            },
            Action: func(ctx *cli.Context) error {
              c := client(ctx)
              result, err := c.Dump(context.Background(), connect.NewRequest(&v1.DumpRequest{}))
              if err != nil {
                return err
              }
              file, err := os.OpenFile(ctx.String("file"), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
              if err != nil {
                return err
              }
              defer file.Close()
              if _, err = file.WriteString(result.Msg.Dump); err != nil {
                return err
              }
              return nil
            },
          },
          {
            Name:  "restore",
            Usage: "load the whole ipam db from json file, previously created, only works if database is already empty",
            Flags: []cli.Flag{
              &cli.StringFlag{
                Name: "file",
              },
            },
            Action: func(ctx *cli.Context) error {
              c := client(ctx)
              json, err := os.ReadFile(ctx.String("file"))
              if err != nil {
                return err
              }
              _, err = c.Load(context.Background(), connect.NewRequest(&v1.LoadRequest{
                Dump: string(json),
              }))
              
              if err != nil {
                return err
              }
              fmt.Printf("database restored\n")
              return nil
            },
          },
        },
      },
    },
  }
  err := app.Run(os.Args)
  if err != nil {
    fmt.Println(err.Error())
    os.Exit(1)
  }
}

func client(ctx *cli.Context) apiv1connect.IpamServiceClient {
  clientOpts, _ := compress.All(compress.LevelBalanced)
  
  return apiv1connect.NewIpamServiceClient(
    http.DefaultClient,
    ctx.String("grpc-server-endpoint"),
    connect.WithGRPC(),
    clientOpts,
  )
}
