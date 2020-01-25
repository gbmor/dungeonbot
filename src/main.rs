use irc::client::prelude::*;

#[macro_use]
extern crate lazy_static;

type Result<T> = std::result::Result<T, irc::error::IrcError>;

lazy_static! {
    static ref CHANNELS: Vec<String> = vec!["#d&d".to_string()];
}
static NICK: &str = "dungeonbot";
static SERVER: &str = "irc.tilde.chat";
static PORT: u16 = 6697;
static USE_SSL: bool = true;

#[cfg_attr(tarpaulin, skip)]
fn main() -> Result<()> {
    println!("dungeonbot 0.1");
    println!();

    let config = Config {
        nickname: Some(NICK.to_string()),
        server: Some(SERVER.to_string()),
        port: Some(PORT),
        use_ssl: Some(USE_SSL),
        channels: Some((*CHANNELS).clone()),
        ..Config::default()
    };

    let client = IrcClient::from_config(config)?;
    client.identify()?;

    client.for_each_incoming(|ircmsg| {
        if let Command::PRIVMSG(chan, msg) = ircmsg.command {
            eprintln!("{}", msg);
        }
    })?;

    Ok(())
}
