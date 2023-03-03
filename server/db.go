package server

import (
	"context"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/google/uuid"
	"github.com/guregu/dynamo"
)

const (
	// usersTableBaseName is the base name of the users table. The base name is
	// prepended with the table prefix to get the full table name.
	usersTableBaseName = "users"
	// tokensTableBaseName is the base name of the tokens table. The base name is
	// prepended with the table prefix to get the full table name.
	tokensTableBaseName = "tokens"
	// eatsTableBaseName is the base name of the eats table. The base name is
	// prepended with the table prefix to get the full table name.
	eatsTableBaseName = "eats"
	// poopsTableBaseName is the base name of the poops table. The base name is
	// prepended with the table prefix to get the full table name.
	poopsTableBaseName = "poops"
	// foodsTableBaseName is the base name of the foods table. The base name is
	// prepended with the table prefix to get the full table name.
	foodsTableBaseName = "foods"
	// foodNamesTableBaseName is the base name of the food names table. The base
	// name is prepended with the table prefix to get the full table name.
	foodNamesTableBaseName = "food_names"
	// ingredientsTableBaseName is the base name of the ingredients table. The
	// base name is prepended with the table prefix to get the full table name.
	ingredientsTableBaseName = "ingredients"
)

// User is a user of Potty Trainer.
type User struct {
	// ID is the unique identifier for this user. It is an opaque string.
	ID string `dynamo:"id,hash"`
	// CreatedAt is the time this user record was created.
	CreatedAt time.Time `dynamo:"created_at"`
	// UpdatedAt is the time this user record was last updated.
	UpdatedAt time.Time `dynamo:"updated_at"`
}

// APIToken wraps a valid API token for a user.
type APIToken struct {
	// ID is the unique identifier for this token. It is not the token itself. ID
	// is an opaque string.
	ID string `dynamo:"id,hash"`
	// UserID is the ID of the user this token belongs to.
	UserID string `dynamo:"user_id"`
	// Token is the actual token. It is an opaque string.
	Token string `dynamo:"token,range"`
	// CreatedAt is the time this token record was created.
	CreatedAt time.Time `dynamo:"created_at"`
	// UpdatedAt is the time this token record was last updated.
	UpdatedAt time.Time `dynamo:"updated_at"`
}

// Eat is a single occurrence of eating. Each eat contains exactly one food
// item. If multiple food items are eaten at the same time, each is their own
// Eat.
type Eat struct {
	// ID is the unique identifier for this eat. It is an opaque string.
	ID string `dynamo:"id,hash"`
	// UserID is the ID of the user who ate the food.
	UserID string `dynamo:"user_id,range"`
	// FoodID is the ID of the food item that was eaten.
	FoodID string `dynamo:"food_id"`
	// FoodText is the exact text of the food item that was eaten. This is saved
	// for forwards compatibility with changes to the Food table or the
	// introduction of new aliases or ingredient relationships.
	FoodText string `dynamo:"food_text"`
	// AteAt is the time at which the food was eaten.
	AteAt time.Time `dynamo:"ate_at"`
	// CreatedAt is the time at which the Eat record was created.
	CreatedAt time.Time `dynamo:"created_at"`
	// UpdatedAt is the time at which the Eat was record last updated.
	UpdatedAt time.Time `dynamo:"updated_at"`
}

// Poop is a single occurrence of pooping.
type Poop struct {
	// ID is the unique identifier for this poop. It is an opaque string.
	ID string `dynamo:"id,hash"`
	// UserID is the ID of the user who pooped.
	UserID string `dynamo:"user_id,range"`
	// PoopedAt is the time at which the poop occurred.
	PoopedAt time.Time `dynamo:"pooped_at"`
	// Quality is the user-reported quality of the poop. This is either -1 or 1;
	// it is an integer for forwards compatibility reasons.
	Quality int `dynamo:"quality"`
	// CreatedAt is the time at which the Poop record was created.
	CreatedAt time.Time `dynamo:"created_at"`
	// UpdatedAt is the time at which the Poop record was last updated.
	UpdatedAt time.Time `dynamo:"updated_at"`
}

// Food is a food item that can be eaten. Foods can be associated with other
// foods through aliases and ingredients.
//
// Foods do not have inherent names. Instead, they are associated with a set of
// names via FoodName records.
//
// Foods are not shared between users. Each user has their own set of foods.
type Food struct {
	// ID is the unique identifier for this food item. It is an opaque string.
	ID string `dynamo:"id,hash"`
	// UserID is the ID of the user who owns this food item. Food items are not
	// shared between users. For example, if two users each eat "yogurt", each Eat
	// record will be associated with its own Food record called "yogurt".
	UserID string `dynamo:"user_id,range"`
	// CreatedAt is the time at which the Food record was created.
	CreatedAt time.Time `dynamo:"created_at"`
	// UpdatedAt is the time at which the Food record was last updated.
	UpdatedAt time.Time `dynamo:"updated_at"`
}

// FoodName is one possible name for a food item.
//
// For example, if a food item is associated with the names "yogurt" and
// "yoghurt", then a user who submits that they have eaten either yogurt or
// yoghurt (or a mix of both) will be considered to have eaten one Food.
type FoodName struct {
	// ID is the unique identifier for this food name. It is an opaque string.
	ID string `dynamo:"id,hash"`
	// UserID is the ID of the user who owns this food name. Food names are not
	// shared between users.
	UserID string `dynamo:"user_id,range"`
	// FoodID is the ID of the food item that this name represents.
	FoodID string `dynamo:"food_id"`
	// Name is the name of the food item pointed to by FoodID, as represented by
	// this record.
	Name string `dynamo:"name"`
	// CreatedAt is the time at which the FoodName record was created.
	CreatedAt time.Time `dynamo:"created_at"`
	// UpdatedAt is the time at which the FoodName record was last updated.
	UpdatedAt time.Time `dynamo:"updated_at"`
}

// Ingredient is an "x contains y" relationship between two food items. This is
// a many-to-many relationship.
//
// For example, "caffeine" is an ingredient of "coffee", and "coffee" and "eggs"
// are ingredients of "coffee cake". Therefore, a person who eats coffee cake
// will be considered to have eaten four items: caffeine, coffee, eggs, and
// coffee cake.
//
// Ingredient relationships may be purely semantic. For example, "fruit" may be
// considered an ingredient of "apple", as every time a person eats an apple
// they are also eating fruit. This is different from an alias relationship,
// where only one food is considered to have been eaten.
type Ingredient struct {
	// ID is the unique identifier for this ingredient relationship. It is an
	// opaque string.
	ID string `dynamo:"id,hash`
	// UserID is the ID of the user who owns this ingredient relationship.
	// Ingredient relationships are not shared between users.
	UserID string `dynamo:"user_id,range"`
	// ResultingFoodID is the ID of the food item that contains the ingredient.
	ResultingFoodID string `dynamo:"resulting_food_id"`
	// ComponentFoodID is the ID of the food item that is the ingredient.
	ComponentFoodID string `dynamo:"component_food_id"`
	// CreatedAt is the time at which the Ingredient record was created.
	CreatedAt time.Time `dynamo:"created_at"`
	// UpdatedAt is the time at which the Ingredient record was last updated.
	UpdatedAt time.Time `dynamo:"updated_at"`
}

type DB struct {
	// db is the underlying DynamoDB client.
	db *dynamo.DB
	// users is the DynamoDB table that holds user records.
	users dynamo.Table
	// tokens is the DynamoDB table that holds API token records.
	tokens dynamo.Table
	// eats is the DynamoDB table that holds eat records.
	eats dynamo.Table
	// poops is the DynamoDB table that holds poop records.
	poops dynamo.Table
	// foods is the DynamoDB table that holds food records.
	foods dynamo.Table
	// foodNames is the DynamoDB table that holds food name records.
	foodNames dynamo.Table
	// ingredients is the DynamoDB table that holds ingredient records.
	ingredients dynamo.Table
}

type DBConfig struct {
	// Region is the AWS region in which the DynamoDB tables are located.
	Region string
	// Endpoint is the endpoint to use when connecting to DynamoDB. This is
	// useful for testing against a local DynamoDB instance.
	Endpoint string
	// TableNamePrefix is the prefix to use for all table names. If this does not
	// end in "-" or "_", a hyphen will added.
	TableNamePrefix string
}

// table returns the DynamoDB table for the given name and configuration. If the
// table does not exist, it is created according to the given schema, which
// should be
func (db *DB) table(cfg *DBConfig, name string, schema any) (dynamo.Table, error) {
	tableName := cfg.TableNamePrefix + name
	t := db.db.Table(tableName)
	_, err := t.Describe().Run()
	if err == nil {
		return t, nil
	}

	if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() != dynamodb.ErrCodeResourceNotFoundException {
		return dynamo.Table{}, err
	}
	if err := db.db.CreateTable(tableName, User{}).Run(); err != nil {
		return dynamo.Table{}, err
	}
	return db.db.Table(tableName), nil
}

func NewDB(cfg *DBConfig) (*DB, error) {
	if cfg == nil {
		cfg = &DBConfig{}
	}
	if cfg.Region == "" {
		cfg.Region = "us-east-1"
	}
	if cfg.Endpoint == "" {
		cfg.TableNamePrefix = "pottytrainer"
	}

	if !strings.HasSuffix(cfg.TableNamePrefix, "-") && !strings.HasSuffix(
		cfg.TableNamePrefix, "_") {
		cfg.TableNamePrefix += "-"
	}

	sess := session.Must(session.NewSession())
	var db DB
	db.db = dynamo.New(sess, &aws.Config{
		Region: aws.String(cfg.Region),
	})

	var err error
	if db.users, err = db.table(cfg, usersTableBaseName, User{}); err != nil {
		return nil, err
	}
	if db.tokens, err = db.table(cfg, tokensTableBaseName, APIToken{}); err != nil {
		return nil, err
	}
	if db.eats, err = db.table(cfg, eatsTableBaseName, Eat{}); err != nil {
		return nil, err
	}
	if db.poops, err = db.table(cfg, poopsTableBaseName, Poop{}); err != nil {
		return nil, err
	}
	if db.foods, err = db.table(cfg, foodsTableBaseName, Food{}); err != nil {
		return nil, err
	}
	if db.foodNames, err = db.table(cfg, foodNamesTableBaseName, FoodName{}); err != nil {
		return nil, err
	}
	if db.ingredients, err = db.table(cfg, ingredientsTableBaseName, Ingredient{}); err != nil {
		return nil, err
	}

	return &db, nil
}

func (db *DB) userFromToken(ctx context.Context, token string) (*User, error) {
	var apiToken APIToken
	if err := db.tokens.Get("user_id", token).OneWithContext(ctx, &apiToken); err != nil {
		return nil, err
	}
	var user User
	if err := db.users.Get("id", apiToken.UserID).OneWithContext(ctx, &user); err != nil {
		return nil, err
	}
	return &user, nil
}

func (db *DB) logPoop(ctx context.Context, userID string, poopedAt time.Time, bad bool) error {
	db.poops.Put(&Poop{
		ID:        uuid.New().String(),
		UserID:    "test",
		PoopedAt:  poopedAt,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}).Run()
	return nil
}
