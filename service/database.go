package service

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/hbahadorzadeh/key-master/util"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"gopkg.in/errgo.v2/fmt/errors"
)

type MongoDB struct {
	client   *mongo.Client
	dbname   string
	logger   *log.Logger
	validate *validator.Validate
}

func NewMongoDatabase(config *util.Configs, logger *log.Logger, validate *validator.Validate) *MongoDB {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().SetAppName("key-master")

	hosts := make([]string, 0)
	logger.Infof("DB Address: `%`", config.DB.MongoServers)
	for _, host := range config.DB.MongoServers {
		hosts = append(hosts, fmt.Sprintf("%s:%d", host.MongoHost, host.MongoPort))
	}
	clientOptions.SetHosts(hosts)
	if config.DB.MongoAuth.Username != "" {
		clientOptions.SetAuth(
			options.Credential{
				AuthMechanism:           config.DB.MongoAuth.AuthMechanism,
				AuthMechanismProperties: config.DB.MongoAuth.AuthMechanismProperties,
				AuthSource:              config.DB.MongoAuth.AuthSource,
				Username:                config.DB.MongoAuth.Username,
				Password:                config.DB.MongoAuth.Password,
			})
	}
	var maxPoolSize = 20
	if config.DB.PoolSize != 0 {
		maxPoolSize = config.DB.PoolSize
	}
	clientOptions.SetMaxPoolSize(uint64(maxPoolSize))
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		logger.Error(err)
	}

	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		logger.Error(err)
	}

	return &MongoDB{
		client:   client,
		dbname:   config.DB.MongoDB,
		logger:   logger,
		validate: validate,
	}
}

func (mdb MongoDB) database() *mongo.Database {
	return mdb.client.Database(mdb.dbname)
}
func (mdb *MongoDB) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return mdb.client.Disconnect(ctx)
}

func (mdb *MongoDB) GetCollection(model interface{}) string {
	modelName := strings.Split(reflect.TypeOf(model).String(), ".")
	return modelName[len(modelName)-1]
}

func (mdb *MongoDB) CreateCollection(model interface{}) {
	collection := mdb.GetCollection(model)
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	query := bson.D{{"name", collection}}
	res, err := mdb.database().ListCollectionNames(ctx, query)
	if err != nil {
		mdb.logger.Error(err)
	}
	if len(res) == 0 {
		err := mdb.database().CreateCollection(ctx, collection)
		if err != nil {
			mdb.logger.Error(err)
		}
	}
}

func (mdb *MongoDB) Select(model interface{}, changes bson.M) error {
}

func (mdb *MongoDB) Create(model interface{}) error {
	err := mdb.validate.Struct(model)
	if err != nil {
		mdb.logger.Error(err)
	}

	collection := mdb.GetCollection(model)

	setBasicData(model, CreatedAt, primitive.NewDateTimeFromTime(time.Now()))
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	res, err := mdb.database().Collection(collection).InsertOne(ctx, model)
	if err != nil {
		return err
	}
	setBasicData(model, ID, res.InsertedID)
	return nil

}

func (mdb *MongoDB) Update(model interface{}, changes bson.M) error {
	err := mdb.validate.Struct(model)
	if err != nil {
		mdb.logger.Error(err)
	}

	collection := mdb.GetCollection(model)

	setBasicData(model, UpdatedAt, primitive.NewDateTimeFromTime(time.Now()))
	id := getBasicDataID(model)
	if id == primitive.NilObjectID || id.String() == "" {
		return errors.New("ID is not set")
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	_, err = mdb.database().Collection(collection).UpdateByID(ctx, bson.M{"_id": id}, changes)
	return err
}

func (mdb *MongoDB) Set(model interface{}) error {
	err := mdb.validate.Struct(model)
	if err != nil {
		mdb.logger.Error(err)
	}

	collection := mdb.GetCollection(model)

	setBasicData(model, UpdatedAt, primitive.NewDateTimeFromTime(time.Now()))
	id := getBasicDataID(model)
	if id == primitive.NilObjectID || id.String() == "" {
		return errors.New("ID is not set")
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	_, err = mdb.database().Collection(collection).ReplaceOne(ctx, bson.M{"_id": id}, model)
	return err
}

func (mdb *MongoDB) Delete(model interface{}) error {
	err := mdb.validate.Struct(model)
	if err != nil {
		mdb.logger.Error(err)
	}

	collection := mdb.GetCollection(model)

	setBasicData(model, DeletedAt, primitive.NewDateTimeFromTime(time.Now()))
	id := getBasicDataID(model)
	if id == primitive.NilObjectID || id.String() == "" {
		return errors.New("ID is not set")
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	_, err = mdb.database().Collection(collection).ReplaceOne(ctx, bson.M{"_id": id}, model)
	return err
}

type BasicDataField string

const (
	ID        BasicDataField = "ID"
	CreatedAt BasicDataField = "CreatedAt"
	UpdatedAt BasicDataField = "UpdatedAt"
	DeletedAt BasicDataField = "DeletedAt"
)

type BasicData struct {
	ID        primitive.ObjectID `json:"id" bson:"-"`
	CreatedAt primitive.DateTime `json:"created_at" bson:"created_at"`
	UpdatedAt primitive.DateTime `json:"updated_at" bson:"updated_at"`
	DeletedAt primitive.DateTime `json:"deleted_at" bson:"deleted_at"`
}

func setBasicData(model interface{}, field BasicDataField, value interface{}) error {
	return setModelField(model, string(field), value)
}

func setModelField(model interface{}, field string, value interface{}) error {
	// pointer to struct - addressable
	ps := reflect.ValueOf(model)
	// struct
	s := ps.Elem()
	if s.Kind() == reflect.Struct {
		// exported field
		f := s.FieldByName(field)
		if f.IsValid() {
			// A Value can be changed only if it is
			// addressable and was not obtained by
			// the use of unexported struct fields.
			if f.CanSet() {
				// change value of N
				f.Set(reflect.ValueOf(value))
				return nil
			} else {
				return errors.New("Value could not be set")
			}
		} else {

			return errors.New("Field is not valid")
		}
	} else {
		return errors.New("Model is not struct")
	}
}

func getBasicDataID(model interface{}) primitive.ObjectID {
	v, err := getBasicData(model, ID)
	if err != nil {
		return primitive.NilObjectID
	}
	return v.(primitive.ObjectID)
}

func getBasicDataDates(model interface{}, field BasicDataField) primitive.DateTime {
	v, err := getBasicData(model, field)
	if err != nil {
		return 0
	}
	return v.(primitive.DateTime)
}

func getBasicData(model interface{}, field BasicDataField) (interface{}, error) {
	v, err := getModelField(model, string(field))
	if err != nil {
		return nil, err
	}
	return v, nil
}

func getModelField(model interface{}, field string) (interface{}, error) {
	// pointer to struct - addressable
	ps := reflect.ValueOf(model)
	// struct
	s := ps.Elem()
	if s.Kind() == reflect.Struct {
		// exported field
		f := s.FieldByName(field)
		if f.IsValid() {
			// A Value can be changed only if it is
			// addressable and was not obtained by
			// the use of unexported struct fields.
			k := f.Kind()
			switch k {
			case reflect.Int64:
				return f.Int(), nil
			case reflect.String:
				return f.String(), nil
			case reflect.Array:
				return f.Interface(), nil
			default:
				return nil, errors.New("Type not supported")
			}
		} else {
			return nil, errors.New("Field is not valid")
		}
	} else {
		return nil, errors.New("Model is not struct")
	}
}
