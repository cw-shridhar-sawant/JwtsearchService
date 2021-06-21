package main

import (
	"context"
	pb "github.com/Cloudwalker-Technologies/JwtsearchService/proto"
	"github.com/golang/protobuf/jsonpb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	//"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
)

type jwserver struct {
	contentCollection *mongo.Collection
}

func (j *jwserver) Getjwtdetails(ctx context.Context, info *pb.TvInfo) (*pb.Tile, error) {
	//some important values are missing get it from header
	log.Println(info)
	resp := new(pb.Tile)
	/*err , tvInfo := makeTvInfo(ctx)
	if err != nil {
		return nil, err
	}
	log.Println('INFORMATION DETAILS', tvInfo)

	 */
	log.Println("Rows config hit with genre and Sources ", info.GetGenres(),info.GetSources())

	tilequery := bson.D{
		{
			Key:   "genre_ids",
			Value: info.GetGenres(),
		},
		{
			Key: "sources",
			Value: info.GetSources(),
		},
	}
	result := j.contentCollection.FindOne(ctx, tilequery)
	if result.Err() != nil {
		if result.Err() == mongo.ErrNoDocuments {
			return nil, status.Error(codes.FailedPrecondition, "ROW  Documents not found.")
		}
		return nil, status.Error(codes.Internal,  result.Err().Error())
	}

	trendBytes, err := result.DecodeBytes()
	if err != nil {
		return nil, status.Error(codes.Internal, "Error decoding document")
	}
	log.Println(trendBytes)

	err = result.Decode(resp)
	//err = trendBytes.Decode(&resp)

	if err != nil {
		return nil ,err
	}

	return resp, nil
}



func (j *jwserver) Postjwtdetails(ctx context.Context, info *pb.TvInfo) (*pb.Row, error) {
	//some important values are missing get it from header
	log.Println(info)
	resp := new(pb.Row)
	//filterArray := primitive.A{}

	log.Println("Rows config hit with genre and Sources ", info.GetGenres(),info.GetSources(),info.GetCategories())

	/*
	//query
	if len(info.GetQuery()) > 0 {
		titleM := bson.M{
				"query": info.GetQuery(),
				"path":  "title",
			}
		filterArray = append(filterArray, titleM)
	}

	//categoires
	if len(info.GetCategories()) > 0 && info.GetCategories()[0] != "" {
		for _, s := range info.GetCategories() {
			println("GOT CATEGORIE ", s)
			categoriesM := bson.M{
					"query": s,
					"path":  "object_type",
			}
			filterArray = append(filterArray, categoriesM)
		}
	}

	/*
	tilequery := bson.D{
		{
			Key:   "genre_ids",
			Value: info.GetGenres(),
		},
		{
			Key: "sources",
			Value: info.GetSources(),
		},
	}
	 */

	stages := mongo.Pipeline{}
	stage1 := bson.D{
		{
			Key : "$match",
			Value: bson.D{
				{
					Key:  "object_type",
					Value: bson.M{"$in": info.GetCategories()},
				},
			},
		},
	}
	stages = append(stages,stage1)
	//shuffling with random tiles Id Stage
	stage2 := bson.D{
		{
			Key : "$sample",
			Value: bson.M{
				"size" :10 ,
			},
		},
	}
	stages =append(stages,stage2)

	cur, err := j.contentCollection.Aggregate(ctx, mongo.Pipeline{stage1,stage2})
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, status.Error(codes.FailedPrecondition, "No Suggestion for this tile")
		}
		return nil, status.Errorf(codes.Internal, "Error while getting suggestions from DB\n %s", err.Error())
	}
	defer cur.Close(ctx)

	searchResult := new([]*pb.Tile)

	for cur.Next(ctx) {
		tile := new(pb.Tile)
		err = jsonpb.UnmarshalString(cur.Current.String(),tile)
		if err != nil {
			return nil, status.Error(codes.Internal, "Error decoding document")
		}

		//err = cur.Decode(&tile)
		//if err != nil {
		//	return nil, status.Error(codes.Internal, "Error decoding document")
		//}
		log.Println(cur)
		*searchResult = append(*searchResult, tile)
	}

	resp.Tiles = *searchResult
	return resp, nil
}

/*
func makeTvInfo(ctx context.Context) (error, *pb.TvInfo) {
	//check in the incoming context the values
	tvInfo := new(pb.TvInfo)
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return status.Error(codes.DataLoss, "Headers or request body data not present"), nil
	}
	for k, v := range md {
		switch k {
		case "emac":
			tvInfo.Emac = v[0]
			break
		case "wmac":
			tvInfo.Wmac = v[0]
			break
		case "mboard":
			tvInfo.Board = v[0]
			break
		case "panel":
			tvInfo.Panel = v[0]
			break
		case "model":
			tvInfo.Model = v[0]
			break
		case "cotaversion":
			tvInfo.Cota = v[0]
			break
		case "fotaversion":
			tvInfo.Fota = v[0]
			break
		case "brand":
			tvInfo.Brand = v[0]
			break
		case "vendor":
			tvInfo.Vendor = v[0]
			break
		}
	}
	return nil, tvInfo
}
*/