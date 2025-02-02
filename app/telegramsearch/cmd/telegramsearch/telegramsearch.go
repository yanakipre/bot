package main

import (
	_ "embed"
	"github.com/yanakipe/bot/app/telegramsearch/cmd/telegramsearch/internal"
)

func main() {
	internal.Execute()
	return
	//flag.Parse()
	//if flagGenconfig != "" {
	//	staticCfg := staticconfig.Config{}
	//	staticCfg.DefaultConfig()
	//	config.Write(staticCfg, staticConfigFilename+".sample")
	//
	//	if err := staticCfg.Validate(); err != nil {
	//		fmt.Printf("static config validation error: %v\n", err)
	//		os.Exit(1)
	//	}
	//}
	//
	//// Create context that listens for the interrupt signal from the OS.
	//ctx, defaultBehaviourForSignals := signal.NotifyContext(
	//	context.Background(),
	//	syscall.SIGINT,
	//	syscall.SIGTERM,
	//)
	//defer defaultBehaviourForSignals()
	//
	//// parse configs
	//var staticConfig staticconfig.Config
	//if err := config.Load(ctx, "telegramsearch", &staticConfig, staticConfigFilename,
	//	&staticconfig.Backend{},
	//); err != nil {
	//	fmt.Printf("cannot load config %s\n", err)
	//	os.Exit(1)
	//}
	//
	//// set up logger
	//logger.SetNewGlobalLoggerOnce(staticConfig.Logging)
	//defer logger.Close()
	//
	//log := logger.FromContext(ctx)
	//
	//openai := httpopenaiclient.NewClient(staticConfig.OpenAI)
	//
	//// Create an EmbeddingRequest for the user query
	//queryResponse, err := openai.CreateEmbeddings(ctx, openaimodels.ReqCreateEmbeddings{
	//	Input: []string{"How many chucks would a woodchuck chuck"},
	//})
	//if err != nil {
	//	log.Fatal("Error creating query embedding:", zap.Error(err))
	//}
	//
	//// Create an embedding for the target text
	//targetResponse, err := openai.CreateEmbeddings(ctx, openaimodels.ReqCreateEmbeddings{
	//	Input: []string{"How many chucks would a woodchuck chuck if the woodchuck could chuck wood"},
	//})
	//if err != nil {
	//	log.Fatal("Error creating target embedding:", zap.Error(err))
	//}
	//
	//// Now that we have the embeddings for the user query and the target text, we
	//// can calculate their similarity.
	//queryEmbedding := queryResponse.Embeddings[0]
	//targetEmbedding := targetResponse.Embeddings[0]
	//
	//similarity, err := queryEmbedding.DotProduct(&targetEmbedding)
	//if err != nil {
	//	log.Fatal("Error calculating dot product:", zap.Error(err))
	//}
	//
	//log.Info("The similarity score between the query and the target", zap.Float32("sim", similarity))
	//
	//storageRW := postgres.New(staticConfig.PostgresRW)
	//if err := storageRW.Ready(ctx); err != nil {
	//	log.Fatal("Error creating postgres reader", zap.Error(err))
	//}
	//defer closer.Close(ctx, storageRW)
	////
	////_, err = storageRW.CreateChat(ctx, storagemodels.ReqCreateChat{ChatID: "limassol"})
	////if err != nil {
	////	log.Fatal("Error creating chat", zap.Error(err))
	////}
	//
	//ctl, err := controllerv1.New(openai, storageRW)
	//if err != nil {
	//	log.Fatal("Error creating controller", zap.Error(err))
	//}
	//_, err = ctl.DumpChatHistory(ctx, controllerv1models.ReqDumpChatHistory{ChatHistory: cylimassol})
	//if err != nil {
	//	log.Fatal("Error dumping chat history", zap.Error(err))
	//}
}
