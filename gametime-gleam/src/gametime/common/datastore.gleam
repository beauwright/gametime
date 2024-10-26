import bison/bson
import gametime/web
import gleam/dict
import gleam/erlang
import gleam/erlang/process
import gleam/option
import mungo
import mungo/aggregation.{
  Let, add_fields, aggregate, match, pipelined_lookup, to_cursor, unwind,
}
import mungo/client
import mungo/crud.{Sort, Upsert}
import mungo/error

pub type GameRoom {
  CreatedBy(String)
  // NOTE: CreatedBy should be the device sessionID, NOT the playerID
  PlayerState(dict.Dict(String, Player))
  ActivePlayer(option.Option(Player))
  InitialPlayerState(dict.Dict(Int, Player))
}

pub type ActivePlayer {
  PlayerID(Int)
  TurnStart(Int)
}

pub type Player {
  ID(String)
  Name(String)
  Increment(Int)
  InitialTime(Int)
  CurrentTime(Int)
  Position(Int)
}

pub fn create_room(room: GameRoom) -> Nil {
  todo
}

pub fn get_room(
  context: web.Context,
  room_id: String,
) -> Result(option.Option(GameRoom), error.Error) {
  let raw =
    context.mongo_client
    |> mungo.collection("rooms")
    |> mungo.find_one([], [], 10)

  todo
}
//  let rooms =
//    client
//    |> mungo.collection("rooms")
//
//  let _ =
//    users
//    |> mungo.insert_many(
//      [
//        [
//          #("username", bson.String("jmorrow")),
//          #("name", bson.String("vincent freeman")),
//          #("email", bson.String("jmorrow@gattaca.eu")),
//          #("age", bson.Int32(32)),
//        ],
//        [
//          #("username", bson.String("real-jerome")),
//          #("name", bson.String("jerome eugene morrow")),
//          #("email", bson.String("real-jerome@running.at")),
//          #("age", bson.Int32(32)),
//        ],
//      ],
//      128,
//    )
//
//  let _ =
//    users
//    |> mungo.update_one(
//      [#("username", bson.String("real-jerome"))],
//      [
//        #(
//          "$set",
//          bson.Document([
//            #("username", bson.String("eugene")),
//            #("email", bson.String("eugene@running.at ")),
//          ]),
//        ),
//      ],
//      [Upsert],
//      128,
//    )
//
//  let assert Ok(yahoo_cursor) =
//    users
//    |> mungo.find_many(
//      [#("email", bson.Regex(#("yahoo", "")))],
//      [Sort([#("username", bson.Int32(-1))])],
//      128,
//    )
//  let _yahoo_users = mungo.to_list(yahoo_cursor, 128)
//
//  let assert Ok(underage_lindsey_cursor) =
//    users
//    |> aggregate([Let([#("minimum_age", bson.Int32(21))])], 128)
//    |> match([
//      #(
//        "$expr",
//        bson.Document([
//          #(
//            "$lt",
//            bson.Array([bson.String("$age"), bson.String("$$minimum_age")]),
//          ),
//        ]),
//      ),
//    ])
//    |> add_fields([
//      #(
//        "first_name",
//        bson.Document([
//          #(
//            "$arrayElemAt",
//            bson.Array([
//              bson.Document([
//                #(
//                  "$split",
//                  bson.Array([bson.String("$name"), bson.String(" ")]),
//                ),
//              ]),
//              bson.Int32(0),
//            ]),
//          ),
//        ]),
//      ),
//    ])
//    |> match([#("first_name", bson.String("lindsey"))])
//    |> pipelined_lookup(
//      from: "profiles",
//      define: [#("user", bson.String("$username"))],
//      pipeline: [
//        [
//          #(
//            "$match",
//            bson.Document([
//              #(
//                "$expr",
//                bson.Document([
//                  #(
//                    "$eq",
//                    bson.Array([bson.String("$username"), bson.String("$$user")]),
//                  ),
//                ]),
//              ),
//            ]),
//          ),
//        ],
//      ],
//      alias: "profile",
//    )
//    |> unwind("$profile", False)
//    |> to_cursor
//
//  let assert #(option.Some(_underage_lindsey), underage_lindsey_cursor) =
//    underage_lindsey_cursor
//    |> mungo.next(128)
//
//  let assert #(option.None, _) =
//    underage_lindsey_cursor
//    |> mungo.next(128)
//}
//
