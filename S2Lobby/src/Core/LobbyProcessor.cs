using System;
using System.Collections.Concurrent;
using System.Collections.Generic;
using System.Runtime.Remoting.Channels;
using System.Text;
using System.Threading;
using Microsoft.VisualBasic.Logging;
using S2Library.Protocol;

namespace S2Lobby
{
    public class LobbyProcessor : ServerProcessor
    {
        //private readonly byte[] _userData = Crypto.BytesFromHexString("00000000040000000000000000000000000000000000000006000000");
        private readonly byte[] _nicknameData = Crypto.BytesFromHexString("000000003900000000000000000000000000000000000000a2000000785edbc9c8800cd880b8842195a1184c1631e000ff819891094508668e438300886265604788beb8ce644b0c8d0d1c05004a9b0ff3");

        private Server _server;
        private static readonly ConcurrentDictionary<uint, uint> ServerUpdateReceivers = new ConcurrentDictionary<uint, uint>();
        private static readonly ConcurrentDictionary<uint, uint> GlobalChatReceivers = new ConcurrentDictionary<uint, uint>();

        public LobbyProcessor(Program program, uint connection) : base(program, connection)
        {
        }

        public override void Close()
        {
            base.Close();

            if (_server != null)
            {
                Program.Servers.Remove(_server.Id);
                NotifyUnlistServer(_server.Id, _server.Running);
                _server = null;
            }
        }

        protected sealed override bool HandlePayloadType(Payloads.Types payloadType, PayloadPrefix payload, PayloadWriter writer)
        {
            if (base.HandlePayloadType(payloadType, payload, writer))
            {
                return true;
            }

            switch (payloadType)
            {
                case Payloads.Types.RegisterNickname:
                    HandleRegisterNickname((RegisterNickname)payload, writer);
                    return true;
                case Payloads.Types.ConfirmNickname:
                    HandleConfirmNickname((ConfirmNickname)payload, writer);
                    return true;
                case Payloads.Types.GetCDKeys:
                    HandleGetCdKeys((GetCDKeys)payload, writer);
                    return true;
                case Payloads.Types.RequestMOTD:
                    HandleGetMOTD((RequestMOTD)payload, writer);
                    return true;
                case Payloads.Types.GetUserInfo:
                    HandleGetUserInfo((GetUserInfo)payload, writer);
                    return true;
                case Payloads.Types.GetPlayerInfo:
                    HandleGetCharacterInfo((GetPlayerInfo)payload, writer);
                    return true;
                case Payloads.Types.GetChatServer:
                    HandleGetChatServer((GetChatServer)payload, writer);
                    return true;
                case Payloads.Types.SelectNickname:
                    HandleSelectNickname((SelectNickname)payload, writer);
                    return true;
                case Payloads.Types.RegisterServer:
                    HandleRegisterServer((RegisterServer)payload, writer);
                    return true;
                case Payloads.Types.GetServers:
                    HandleGetServers((GetServers)payload, writer);
                    return true;
                case Payloads.Types.StopServerUpdates:
                    HandleStopServerUpdates((StopServerUpdates)payload, writer);
                    return true;
                case Payloads.Types.UnknownType056:
                    HandlePayload056((Payload56)payload, writer);
                    return true;
                case Payloads.Types.RegObserverUserLogin:
                    HandleRegObserverUserLogin((RegObserverUserLogin)payload, writer);
                    return true;
                case Payloads.Types.DeregObserverUserLogin:
                    HandleDeregObserverUserLogin((DeregObserverUserLogin)payload, writer);
                    return true;
                case Payloads.Types.UnknownType157:
                    HandlePayload157((Payload157)payload, writer);
                    return true;
                case Payloads.Types.ConnectToServer:
                    HandleConnectToServer((ConnectToServer)payload, writer);
                    return true;
                case Payloads.Types.UpdateServerInfo:
                    HandleUpdateServerInfo((UpdateServerInfo)payload, writer);
                    return true;
                case Payloads.Types.UnlistServer:
                    HandleUnlistServer((UnlistServer) payload, writer);
                    return true;
                case Payloads.Types.PlayerJoinedServer:
                    HandlePlayerJoinedServer((PlayerJoinedServer)payload, writer);
                    return true;
                case Payloads.Types.PlayerLeftServer:
                    HandlePlayerLeftServer((PlayerLeftServer)payload, writer);
                    return true;
                case Payloads.Types.JoinServer:
                    HandleJoinServer((JoinServer) payload, writer);
                    return true;
                case Payloads.Types.LeaveServer:
                    HandleLeaveServer((LeaveServer) payload, writer);
                    return true;
                
                // Chat related packages
                case Payloads.Types.RegObserverGlobalChat:
                    HandleRegObserverGlobalChat((RegObserverGlobalChat)payload, writer);
                    return true;
                case Payloads.Types.DeregObserverGlobalChat:
                    HandleDeregObserverGlobalChat((DeregObserverGlobalChat)payload, writer);
                    return true;
                case Payloads.Types.ChatPayload:
                    HandleChatPayload((ChatPayload) payload, writer);
                    return true;
                
                
                default:
                    Logger.LogError($"{payloadType} not implemented!");
                    return false;
            }
        }

        private void HandleRegisterNickname(RegisterNickname payload, PayloadWriter writer)
        {
            uint accountId = payload.OwnerId;
            string nickname = payload.Name;

            if (accountId != 0)
            {
                StatusWithId resultPayload1 = Payloads.CreatePayload<StatusWithId>();
                resultPayload1.Errorcode = 1;
                resultPayload1.Errormsg = "Incorrect account";
                resultPayload1.Id = payload.OwnerId;
                resultPayload1.TicketId = payload.TicketId;
                SendReply(writer, resultPayload1);
                return;
            }

            Program.Accounts.SetNickname(Database.Connection, Account.Id, nickname);
            Account.PlayerName = nickname;

            StatusWithId resultPayload2 = Payloads.CreatePayload<StatusWithId>();
            resultPayload2.Errorcode = 0;
            resultPayload2.Errormsg = null;
            resultPayload2.Id = Account.Id;
            resultPayload2.TicketId = payload.TicketId;
            SendReply(writer, resultPayload2);
        }

        private void HandleConfirmNickname(ConfirmNickname payload, PayloadWriter writer)
        {
            uint accountId = payload.UserId;
            if (accountId != Account.Id)
            {
                ResultStatusMsg resultPayload1 = Payloads.CreatePayload<ResultStatusMsg>();
                resultPayload1.Errorcode = 1;
                resultPayload1.Errormsg = "Incorrect account";
                resultPayload1.TicketId = payload.TicketId;
                SendReply(writer, resultPayload1);
                return;
            }

            string mail = payload.Mail;
            if (mail != null)
            {
                Program.Accounts.SetEmail(Database.Connection, accountId, mail);
                Account.Email = mail;
            }

            byte[] nicknameData = payload.Data;
            if (nicknameData != null)
            {
                Program.Accounts.SetUserData(Database.Connection, accountId, nicknameData);
                Account.UserData = nicknameData;
            }

            ResultStatusMsg resultPayload2 = Payloads.CreatePayload<ResultStatusMsg>();
            resultPayload2.Errorcode = 0;
            resultPayload2.Errormsg = null;
            resultPayload2.TicketId = payload.TicketId;
            SendReply(writer, resultPayload2);
        }

        private void HandleGetCdKeys(GetCDKeys payload, PayloadWriter writer)
        {
            uint accountId = payload.UserId;
            if (accountId != Account.Id)
            {
                ResultStatusMsg resultPayload1 = Payloads.CreatePayload<ResultStatusMsg>();
                resultPayload1.Errorcode = 1;
                resultPayload1.Errormsg = "Incorrect account";
                resultPayload1.TicketId = payload.TicketId;
                SendReply(writer, resultPayload1);
                return;
            }

            SendCDKey resultPayload2 = Payloads.CreatePayload<SendCDKey>();
            resultPayload2.UserId = accountId;
            resultPayload2.CdKey = "0000000000000000";
            resultPayload2.Keypool = 1;
            resultPayload2.TicketId = payload.TicketId;
            SendReply(writer, resultPayload2);

            ResultStatusMsg resultPayload3 = Payloads.CreatePayload<ResultStatusMsg>();
            resultPayload3.Errorcode = 0;
            resultPayload3.Errormsg = null;
            resultPayload3.TicketId = payload.TicketId;
            SendReply(writer, resultPayload3);
        }

        private void HandleGetMOTD(RequestMOTD payload, PayloadWriter writer)
        {
            // For whatever reason MOTD is encoded as UTF-8
            byte[] motd = Encoding.UTF8.GetBytes(
                Constants.MOTD.Replace("%name%", Account.GetUserNameStripped()));
            
            SendMOTD resultPayload = Payloads.CreatePayload<SendMOTD>();
            resultPayload.Txt = motd;
            resultPayload.TicketId = payload.TicketId;
            SendReply(writer, resultPayload);
        }

        private void HandleGetUserInfo(GetUserInfo payload, PayloadWriter writer)
        {
            uint accountId = payload.UserId;
            if (accountId != Account.Id)
            {
                ResultStatusMsg resultPayload1 = Payloads.CreatePayload<ResultStatusMsg>();
                resultPayload1.Errorcode = 1;
                resultPayload1.Errormsg = "Incorrect account";
                resultPayload1.TicketId = payload.TicketId;
                SendReply(writer, resultPayload1);
                return;
            }

            SendUserInfo resultPayload2 = Payloads.CreatePayload<SendUserInfo>();
            resultPayload2.UserId = accountId;
            resultPayload2.Name = Account.UserName;
            resultPayload2.Password = null;
            resultPayload2.Mail = Account.Email;
            resultPayload2.Banned = false;
            resultPayload2.Active = true;
            resultPayload2.Status = 2;
            resultPayload2.Data = Account.UserData;
            resultPayload2.Created = "2012-04-02 00:00:00+0:00";
            resultPayload2.LastLogin = "2012-04-02 00:00:00+0:00";
            resultPayload2.TotalLogins = 1;
            resultPayload2.TicketId = payload.TicketId;
            SendReply(writer, resultPayload2);

            ResultStatusMsg resultPayload3 = Payloads.CreatePayload<ResultStatusMsg>();
            resultPayload3.Errorcode = 0;
            resultPayload3.Errormsg = null;
            resultPayload3.TicketId = payload.TicketId;
            SendReply(writer, resultPayload3);
        }

        private void HandleGetCharacterInfo(GetPlayerInfo payload, PayloadWriter writer)
        {
            uint accountId = payload.UserId;
            if (accountId != Account.Id)
            {
                ResultStatusMsg resultPayload1 = Payloads.CreatePayload<ResultStatusMsg>();
                resultPayload1.Errorcode = 1;
                resultPayload1.Errormsg = "Incorrect account";
                resultPayload1.TicketId = payload.TicketId;
                SendReply(writer, resultPayload1);
                return;
            }

            SendPlayerInfo resultPayload2 = Payloads.CreatePayload<SendPlayerInfo>();
            resultPayload2.CharId = Account.Id;
            resultPayload2.Name = Account.PlayerName;
            resultPayload2.OwnerId = Account.Id;
            resultPayload2.OwnerName = Account.UserName;
            resultPayload2.GuildId = 0;
            resultPayload2.GuildName = null;
            resultPayload2.GuildRole = 0;
            resultPayload2.Status = 1;
            resultPayload2.ServerId = 0;
            resultPayload2.ServerName = null;
            resultPayload2.Data = _nicknameData;
            resultPayload2.TicketId = payload.TicketId;
            SendReply(writer, resultPayload2);

            ResultStatusMsg resultPayload3 = Payloads.CreatePayload<ResultStatusMsg>();
            resultPayload3.Errorcode = 0;
            resultPayload3.Errormsg = null;
            resultPayload3.TicketId = payload.TicketId;
            SendReply(writer, resultPayload3);
        }

        private void HandleGetChatServer(GetChatServer payload, PayloadWriter writer)
        {
            SendChatServerInfo resultPayload = Payloads.CreatePayload<SendChatServerInfo>();
            resultPayload.ServerId = 1;
            resultPayload.Ip = Config.Get("chat/ip");
            resultPayload.Port = Config.GetInt("chat/port");
            resultPayload.ServerType = payload.ServerType;
            resultPayload.Version = null;
            resultPayload.Data = null;
            resultPayload.TicketId = payload.TicketId;
            SendReply(writer, resultPayload);
        }

        private void HandleSelectNickname(SelectNickname payload, PayloadWriter writer)
        {
            uint playerId = payload.CharId;

            Account account = Program.Accounts.Get(Database.Connection, playerId);
            if (account == null)
            {
                ResultStatusMsg resultPayload1 = Payloads.CreatePayload<ResultStatusMsg>();
                resultPayload1.Errorcode = 1;
                resultPayload1.Errormsg = "Incorrect account";
                resultPayload1.TicketId = payload.TicketId;
                SendReply(writer, resultPayload1);
                return;
            }

            SelectNicknameReply resultPayload2 = Payloads.CreatePayload<SelectNicknameReply>();
            resultPayload2.CharId = account.Id;
            resultPayload2.Name = account.PlayerName;
            resultPayload2.OwnerId = account.Id;
            resultPayload2.OwnerName = account.UserName;
            resultPayload2.GuildId = 0;
            resultPayload2.GuildName = null;
            resultPayload2.GuildRole = 0;
            resultPayload2.Status = 1;
            resultPayload2.ServerId = 0;
            resultPayload2.ServerName = null;
            resultPayload2.Data = _nicknameData;
            resultPayload2.TicketId = payload.TicketId;
            SendReply(writer, resultPayload2);

            ResultStatusMsg resultPayload3 = Payloads.CreatePayload<ResultStatusMsg>();
            resultPayload3.Errorcode = 0;
            resultPayload3.Errormsg = null;
            resultPayload3.TicketId = payload.TicketId;
            SendReply(writer, resultPayload3);
        }

        private void HandleRegisterServer(RegisterServer payload, PayloadWriter writer)
        {
            // TODO
            uint serverId = Program.Servers.Register(payload.Name);
            if (serverId == 0)
            {
                // TODO implement error codes if available
                SendReply(writer, Payloads.CreateStatusFailMsg("Failed to register server", payload.TicketId));
                return;
            }

            _server = Program.Servers.Get(serverId);
            if (_server == null)
            {
                Program.Servers.Remove(serverId);
                SendReply(writer, Payloads.CreateStatusFailMsg("Failed to register server", payload.TicketId));
                return;
            }

            _server.Players.TryAdd(Account.Id, Connection);
            
            _server.ConnectionId = Connection;
            _server.OwnerId = Account.Id;
            _server.Description = payload.Description;
            //_server.Ip = "192.168.8.20"; // TODO get ip from connection if possible
            _server.Ip = Config.Get("lobby/ip");
            //_server.Ip = payload.Ip ?? Config.Get("lobby/ip");
            _server.Port = payload.Port;
            // TODO implement remaining
            _server.PlayersTotal = payload.PlayersTotal;
            _server.PlayersAi = 0;
            _server.Map = payload.Map;
            //_server.Map = "MP_2P_Storm_Coast\vfr_11888";
            _server.Running = false;
            
            SendServerUpdates(payload.TicketId);

            var resultPayload = Payloads.CreatePayload<StatusWithId>();
            resultPayload.Errorcode = 0;
            resultPayload.Errormsg = null;
            resultPayload.Id = _server.Id;
            resultPayload.TicketId = payload.TicketId;
            SendReply(writer, resultPayload);
            
            Logger.Log($"User {Account.UserName} created a new lobby as {payload.Name}");
        }
        private void HandleRegisterServerOld(RegisterServer payload, PayloadWriter writer)
        {
            string name = payload.Name;

            uint serverId = Program.Servers.Register(name);
            if (serverId == 0)
            {
                StatusWithId resultPayload1 = Payloads.CreatePayload<StatusWithId>();
                resultPayload1.Errorcode = 3;
                resultPayload1.Errormsg = "Can not register server";
                resultPayload1.Id = 0;
                resultPayload1.TicketId = payload.TicketId;
                SendReply(writer, resultPayload1);
                return;
            }

            _server = Program.Servers.Get(serverId);
            if (_server == null)
            {
                Program.Servers.Remove(serverId);

                StatusWithId resultPayload1 = Payloads.CreatePayload<StatusWithId>();
                resultPayload1.Errorcode = 3;
                resultPayload1.Errormsg = "Can not register server";
                resultPayload1.Id = 0;
                resultPayload1.TicketId = payload.TicketId;
                SendReply(writer, resultPayload1);
                return;
            }

            /*
            _server.ConnectionId = Connection;
            _server.OwnerId = Account.Id;
            _server.Description = payload.Description;
            _server.Ip = payload.Ip ?? Config.Get("lobby/ip");
            _server.Port = payload.Port;
            _server.Type = payload.ServerType;
            _server.SubType = payload.ServerSubtype;
            _server.MaxPlayers = payload.MaxPlayers;
            _server.RoomId = payload.RoomId;
            _server.Level = payload.Level;
            _server.GameMode = payload.GameMode;
            _server.Hardcore = payload.Hardcore;
            _server.Running = payload.Running;
            _server.LockedConfig = payload.LockedConfig;
            _server.Data = payload.Data;
            if (payload.Cipher != null)
            {
                byte[] serverPassword = Crypto.HandleCipher(payload.Cipher, SessionKey);
                int length = System.Array.FindIndex(serverPassword, b => b == 0);
                string password = System.Text.Encoding.ASCII.GetString(serverPassword, 0, length);
                Logger.LogDebug($" Server password is: {password}");

                _server.NeedsPassword = true;
                _server.Password = password;
            }
            */
            SendServerUpdates();

            StatusWithId resultPayload2 = Payloads.CreatePayload<StatusWithId>();
            resultPayload2.Errorcode = 0;
            resultPayload2.Errormsg = null;
            resultPayload2.Id = _server.Id;
            resultPayload2.TicketId = payload.TicketId;
            SendReply(writer, resultPayload2);
        }

        private void HandleGetServers(GetServers payload, PayloadWriter writer)
        {
            if (!ServerUpdateReceivers.TryAdd(Connection, Account.Id))
            {
                ResultStatusMsg resultPayload1 = Payloads.CreatePayload<ResultStatusMsg>();
                resultPayload1.Errorcode = 3;
                resultPayload1.Errormsg = "Can not get server list";
                resultPayload1.TicketId = payload.TicketId;
                SendReply(writer, resultPayload1);
                return;
            }

            List<Server> servers = Program.Servers.GetServers();

            uint ticketId = payload.TicketId;
            foreach (Server server in servers)
            {
                GameServerData resultPayload1 = CreateServerInfoPayload(server, ticketId);
                SendReply(writer, resultPayload1);
            }

            //ServerListTest(writer, payload.TicketId);
            
            ResultStatusMsg resultPayload2 = Payloads.CreatePayload<ResultStatusMsg>();
            resultPayload2.Errorcode = 0;
            resultPayload2.Errormsg = null;
            resultPayload2.TicketId = payload.TicketId;
            SendReply(writer, resultPayload2);
        }

        private void ServerListTest(PayloadWriter writer, uint ticketId)
        {
            GameServerData resultPayload = Payloads.CreatePayload<GameServerData>();

            resultPayload.data = Crypto.BytesFromHexString("05000000 7465737400" +
                                                           "05000000" +
                                                           "01000000 00" +
                                                           "0D000000 3139322E3136382E382E323000" +
                                                           "67150000" +
                                                           "00 00 00 00 00" +
                                                           "06000000313137353700" +
                                                           "06" +
                                                           "01" +
                                                           "04" +
                                                           "00 00 00" +
                                                           "1B000000 4D505F32505F53746F726D5F436F6173740B64655F313137353700" +
                                                           "00" +
                                                           "00 00 00 00" +
                                                           "0A 00 00 00");
                
            resultPayload.ServerId = 54;
            resultPayload.OwnerId = 1;
            resultPayload.Name = "testname";
            resultPayload.Description = "description";
            resultPayload.Ip = "192.168.8.20";
            resultPayload.Port = 5479;
            resultPayload.MaxPlayers = 6;
            resultPayload.CurPlayers = 2;
            resultPayload.AiPlayers = 1;
            resultPayload.Map = "MP_2P_Storm_Coast\vde_11757";
            resultPayload.Running = false;
            resultPayload.TicketId = ticketId;

            SendReply(writer, resultPayload);
        }
        
        private static GameServerData CreateServerInfoPayload(Server server, uint ticketId)
        {
            var resultPayload = Payloads.CreatePayload<GameServerData>();
            resultPayload.ServerId = server.Id;
            resultPayload.OwnerId = server.OwnerId;
            resultPayload.Name = server.Name;
            resultPayload.Description = server.Description;
            resultPayload.Ip = server.Ip;
            resultPayload.Port = server.Port;
            resultPayload.MaxPlayers = server.PlayersTotal;
            resultPayload.CurPlayers = server.GetPlayerCount();
            resultPayload.AiPlayers = server.PlayersAi;
            resultPayload.Map = server.Map;
            resultPayload.Running = server.Running;
            resultPayload.TicketId = ticketId;

            /*
            resultPayload.Unknown0 = 0;
            resultPayload.Unknown1 = 11757;
            resultPayload.Unknown2 = "11757";
            resultPayload.Unknown6 = "11757";
            */
            
            return resultPayload;
        }

        private void HandleStopServerUpdates(StopServerUpdates payload, PayloadWriter writer)
        {
            uint accountId;
            ServerUpdateReceivers.TryRemove(Connection, out accountId);

            SendReply(writer, Payloads.CreateStatusOkMsg(payload.TicketId));
        }

        private void HandlePayload056(Payload56 payload, PayloadWriter writer)
        {
            SendReply(writer, Payloads.CreateStatusOkMsg(payload.TicketId));
        }

        private void HandleRegObserverUserLogin(RegObserverUserLogin payload, PayloadWriter writer)
        {
            //TODO: better implementation?
            if (GlobalLoginReceivers.TryAdd(Connection, Account.Id))
            {
                SendReply(writer, Payloads.CreateStatusOkMsg(payload.TicketId));
            }
            else
            {
                Logger.LogError($"RegObserverUserLogin failed for {Account.UserName}");
            }

            // notify new user about all logged in users
            foreach (KeyValuePair<uint, Account> userOnline in GlobalUsersOnline.ToArray())
            {
                var resultPayload = Payloads.CreatePayload<UserLoggedIn>();
                resultPayload.UserId = userOnline.Value.Id;
                resultPayload.Name = userOnline.Value.UserName;
                SendReply(writer, resultPayload);
            }
        }
        
        private void HandleDeregObserverUserLogin(DeregObserverUserLogin payload, PayloadWriter writer)
        {
            uint accountId;
            if (GlobalLoginReceivers.TryRemove(Connection, out accountId))
            {
                SendReply(writer, Payloads.CreateStatusOkMsg(payload.TicketId));   
            }
            else
            {
                Logger.LogError($"DeregObserverUserLogin failed for {accountId}");
            }
            
            return;

            // notify all users
            foreach (KeyValuePair<uint, uint> loginObserver in GlobalLoginReceivers.ToArray())
            {
                var resultPayload = Payloads.CreatePayload<UserLoggedOut>();
                resultPayload.UserId = accountId;
                SendToLobbyConnection(loginObserver.Key, resultPayload);
            }
        }

        private void HandleJoinServer(JoinServer payload, PayloadWriter writer)
        {
            /*
             * Error IDs:
             * 0x84 (132): GameServer not found
             * 0x87 (135): GameServer full
             */
            Server server = Program.Servers.Get(payload.ServerId);
            if (server == null)
            {
                SendReply(writer, Payloads.CreateStatusFailMsg(
                    0x84, "GameServer could not be found", payload.TicketId));
                return;
            }

            bool full = server.IsFull();
            
            // This might be weird, but I need to do this, because client sends LeaveServer when lobby
            // returns ServerFull error code 
            
            server.Players.TryAdd(payload.UserId, Connection);
            _server = server;
            
            if (full)
            {
                SendReply(writer, Payloads.CreateStatusFailMsg(
                    0x87, "GameServer is already full", payload.TicketId));
            }
            else
            {
                SendReply(writer, Payloads.CreateStatusOkMsg(payload.TicketId));   
            }
        }

        private void HandleLeaveServer(LeaveServer payload, PayloadWriter writer)
        {
            // TODO error messages ?
            if (_server == null)
            {
                SendReply(writer, Payloads.CreateStatusFailMsg("GameServer does not exist", payload.TicketId));
                return;
            }
            uint connection;
            _server.Players.TryRemove(payload.UserId, out connection);
            
            SendReply(writer, Payloads.CreateStatusOkMsg(payload.TicketId));
        }
        
        private void HandlePayload157(Payload157 payload, PayloadWriter writer)
        {
            ResultStatusMsg resultPayload = Payloads.CreatePayload<ResultStatusMsg>();
            resultPayload.Errorcode = 0;
            resultPayload.Errormsg = null;
            resultPayload.TicketId = payload.TicketId;
            SendReply(writer, resultPayload);
        }

        private void HandleConnectToServer(ConnectToServer payload, PayloadWriter writer)
        {
            uint serverId = payload.ServerId;

            Server server = Program.Servers.Get(serverId);
            if (server == null)
            {
                ResultStatusMsg resultPayload = Payloads.CreatePayload<ResultStatusMsg>();
                resultPayload.Errorcode = 1;
                resultPayload.Errormsg = "Unknown server";
                resultPayload.TicketId = payload.TicketId;
                SendReply(writer, resultPayload);
                return;
            }

            byte[] nonce = Crypto.CreateNonce();

            PlayerConnecting resultPayload2 = Payloads.CreatePayload<PlayerConnecting>();
            resultPayload2.Nonce = nonce;
            resultPayload2.CharId = Account.Id;
            resultPayload2.Name = Account.PlayerName;
            resultPayload2.OwnerId = Account.Id;
            resultPayload2.OwnerName = Account.UserName;
            resultPayload2.GuildId = 0;
            resultPayload2.GuildName = null;
            resultPayload2.GuildRole = 0;
            resultPayload2.Data = _nicknameData;
            SendToLobbyConnection(server.ConnectionId, resultPayload2);

            ConnectToServerReply resultPayload3 = Payloads.CreatePayload<ConnectToServerReply>();
            resultPayload3.PermId = Account.Id;
            resultPayload3.ServerId = serverId;
            resultPayload3.Ip = server.Ip;
            resultPayload3.Port = server.Port;
            resultPayload3.Nonce = nonce;
            resultPayload3.Errorcode = 0;
            resultPayload3.Errormsg = null;
            resultPayload3.TicketId = payload.TicketId;
            SendReply(writer, resultPayload3);
        }

        private void HandleUpdateServerInfo(UpdateServerInfo payload, PayloadWriter writer)
        {
            if (_server == null)
            {
                ResultStatusMsg resultPayload1 = Payloads.CreatePayload<ResultStatusMsg>();
                resultPayload1.Errorcode = 3;
                resultPayload1.Errormsg = "No server";
                resultPayload1.TicketId = payload.TicketId;
                SendReply(writer, resultPayload1);
                return;
            }
            
            // TODO
            _server.Name = payload.Name;
            _server.Description = payload.Description;
            _server.PlayersTotal = (byte) (payload.PlayersMax - payload.SlotsOccupied);
            _server.Map = payload.Map;

            SendServerUpdates(payload.TicketId);
            
            SendReply(writer, Payloads.CreateStatusOkMsg(payload.TicketId));
            
            Logger.Log($"Server {payload.Name} got updated by {Account.UserName}");
        }
        private void HandleUpdateServerInfoOld(UpdateServerInfo payload, PayloadWriter writer)
        {
            if (_server == null)
            {
                ResultStatusMsg resultPayload1 = Payloads.CreatePayload<ResultStatusMsg>();
                resultPayload1.Errorcode = 3;
                resultPayload1.Errormsg = "No server";
                resultPayload1.TicketId = payload.TicketId;
                SendReply(writer, resultPayload1);
                return;
            }

            /*
            _server.Name = payload.Name;
            _server.Description = payload.Description;
            _server.MaxPlayers = payload.MaxPlayers;
            _server.RoomId = payload.RoomId;
            _server.Level = payload.Level;
            _server.GameMode = payload.GameMode;
            _server.Hardcore = payload.Hardcore;
            _server.Running = payload.Running;
            _server.LockedConfig = payload.LockedConfig;
            _server.Data = payload.Data;
            if (payload.Cipher == null)
            {
                _server.NeedsPassword = false;
                _server.Password = null;
            }
            else
            {
                byte[] serverPassword = Crypto.HandleCipher(payload.Cipher, SessionKey);
                int length = System.Array.FindIndex(serverPassword, b => b == 0);
                string password = System.Text.Encoding.ASCII.GetString(serverPassword, 0, length);
                Logger.LogDebug($" Server password is: {password}");

                _server.NeedsPassword = true;
                _server.Password = password;
            }
            */

            SendServerUpdates();

            ResultStatusMsg resultPayload2 = Payloads.CreatePayload<ResultStatusMsg>();
            resultPayload2.Errorcode = 0;
            resultPayload2.Errormsg = null;
            resultPayload2.TicketId = payload.TicketId;
            SendReply(writer, resultPayload2);
        }

        private void HandlePlayerJoinedServer(PlayerJoinedServer payload, PayloadWriter writer)
        {
            if (_server == null)
            {
                return;
            }

            _server.Players.TryAdd(payload.PermId, Connection);
            SendServerUpdates();
        }

        private void HandlePlayerLeftServer(PlayerLeftServer payload, PayloadWriter writer)
        {
            if (_server == null)
            {
                return;
            }

            uint dummy;
            _server.Players.TryRemove(payload.PermId, out dummy);
            SendServerUpdates();
        }

        private void HandleUnlistServer(UnlistServer payload, PayloadWriter writer)
        {
            // TODO
            /*
             * TicketIds:
             * 0xB (11): RemoveGameServer
             * 0xE (14): StartGameServer 
             */
            Thread.Sleep(1000);
            //SendReply(writer, Payloads.CreateStatusOkMsg(payload.TicketId));
            
            switch (payload.TicketId)
            {
                case 11:
                    Program.Servers.Remove(payload.ServerId);
                    NotifyUnlistServer(payload.ServerId, payload.Running);

                    if (_server != null && _server.Id == payload.ServerId)
                    {
                        _server = null;
                    }
                    break;
                case 14:
                    Server server = Program.Servers.Get(payload.ServerId);
                    if (server == null)
                    {
                        SendReply(writer, Payloads.CreateStatusFailMsg("ServerId does not exist", payload.TicketId));
                        return;
                    }
                    server.Running = payload.Running;
                    //server.Running = true;

                    var startGameServer = Payloads.CreateStatusOkMsg(payload.TicketId);
                    UnlistServer unlistInfo = Payloads.CreatePayload<UnlistServer>();
                    unlistInfo.ServerId = payload.ServerId;
                    unlistInfo.Running = payload.Running;
                    unlistInfo.TicketId = payload.TicketId;

                    KeyValuePair<uint, uint>[] players = server.Players.ToArray();
                    foreach (KeyValuePair<uint, uint> player in players)
                    {
                        var statusIdOk = Payloads.CreatePayload<StatusWithId>();
                        statusIdOk.Errorcode = 0;
                        statusIdOk.Errormsg = null;
                        statusIdOk.Id = server.Id;
                        statusIdOk.TicketId = payload.TicketId;
                        
                        //SendToLobbyConnection(player.Value, statusIdOk);
                        //SendToLobbyConnection(player.Value, Payloads.CreateStatusOkMsg(payload.TicketId));
                        //SendToLobbyConnection(player.Value, new StartGame());

                        var gameServerData = CreateServerInfoPayload(server, payload.TicketId);
                        SendToLobbyConnection(player.Value, gameServerData);
                    }
                    //SendServerUpdates(payload.TicketId, server);
                    break;
            }
            SendReply(writer, Payloads.CreateStatusOkMsg(payload.TicketId));
        }
        
        private void SendServerUpdates(uint ticketId = 0, Server server = null)
        {
            if (server == null)
            {
                server = _server;   
            }
            GameServerData gameServerData = CreateServerInfoPayload(server, ticketId);
            KeyValuePair<uint, uint>[] servers = ServerUpdateReceivers.ToArray();
            foreach (KeyValuePair<uint, uint> updateRecv in servers)
            {
                SendToLobbyConnection(updateRecv.Key, gameServerData);
            }
        }

        private void NotifyUnlistServer(uint serverId, bool running)
        {
            UnlistServer unlistInfo = Payloads.CreatePayload<UnlistServer>();
            unlistInfo.ServerId = serverId;
            unlistInfo.Running = running;
            KeyValuePair<uint, uint>[] servers = ServerUpdateReceivers.ToArray();
            foreach (KeyValuePair<uint, uint> server in servers)
            {
                SendToLobbyConnection(server.Key, unlistInfo);
            }
        }
        
        // Chat related payloads
        
        private void HandleRegObserverGlobalChat(RegObserverGlobalChat payload, PayloadWriter writer)
        {
            if (GlobalChatReceivers.TryAdd(Connection, Account.Id))
            {
                SendReply(writer, Payloads.CreateStatusOkMsg(payload.TicketId));   
            }
            else
            {
                Logger.LogError($"Failed to add GlobalChatObserver for {Account.Id}");
            }
        }

        private void HandleDeregObserverGlobalChat(DeregObserverGlobalChat payload, PayloadWriter writer)
        {
            uint accountId;
            if (GlobalChatReceivers.TryRemove(Connection, out accountId))
            {
                SendReply(writer, Payloads.CreateStatusOkMsg(payload.TicketId));   
            }
            else
            {
                Logger.LogError($"Failed to DeregObserverChat for {accountId}");
            }
        }
        
        private void HandleChatPayload(ChatPayload payload, PayloadWriter writer)
        {
            // TODO:
            // - chat filter?
            // - proper encoding (ß turns into ?)
            
            // for debugging only
            if (payload.Txt.Contains("s"))
            {
                //ServerListTest(writer, payload.TicketId);
            }
            
            var chatobservers = GlobalChatReceivers.ToArray();
            foreach (KeyValuePair<uint, uint> chatobserver in chatobservers)
            {
                var resultPayload = Payloads.CreatePayload<Chat>();
                resultPayload.Txt = payload.Txt;
                resultPayload.FromId = Account.Id;

                SendToLobbyConnection(chatobserver.Key, resultPayload);
            }
        }
    }
}
