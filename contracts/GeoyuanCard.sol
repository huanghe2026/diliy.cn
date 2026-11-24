// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

/// @title GeoyuanCard — 地理元 ERC-721 卡牌
/// @notice 每张卡绑定一个唯一的 H3 hexagonal grid cell (uint64)
/// @dev   Polygon PoS / Amoy testnet 兼容
contract GeoyuanCard {

    // ─── 状态 ─────────────────────────────────────────────
    string  public constant name     = "GeoyuanCard";
    string  public constant symbol   = "GYC";
    uint256 public          totalSupply;

    // tokenId → 所有者
    mapping(uint256 => address)                       public ownerOf;
    // tokenId → H3 cell index (uint64)
    mapping(uint256 => uint64)                        public cellOf;
    // H3 cell → tokenId（反查，保证每个 cell 只能 mint 一次）
    mapping(uint64  => uint256)                        public cellMinted;
    // tokenId → metadata URI (or JSON blob 直接存链上)
    mapping(uint256 => string)                        public tokenURI;
    // tokenId → metadata 结构体（也存链上，便宜且省一步）
    mapping(uint256 => CardMeta)                      public metadata;
    // 所有者 → tokenId[]（查 MyCards 用）
    mapping(address => mapping(uint256 => uint256))   public tokensOfOwner;
    mapping(address => uint256)                       public balanceOf;

    struct CardMeta {
        string title;
        string description;
        string color;       // hex e.g. "#c77dff"
        uint64 h3Index;    // H3 cell
        uint16 latE4;      // 纬度×10000 (30.2741°N → 302741)
        int16  lonE4;      // 经度×10000 (120.1551°E → 1201551)
        uint32 res;        // H3 resolution
        uint40 timestamp;  // mint 时间戳 (秒)
    }

    // ─── 事件 ─────────────────────────────────────────────
    event Transfer(address indexed from, address indexed to, uint256 indexed tokenId);
    event Minted(uint256 indexed tokenId, address indexed owner, uint64 indexed h3Index);

    // ─── Errors ────────────────────────────────────────────
    error CellAlreadyMinted(uint64 h3Index, uint256 existingTokenId);
    error NotOwner(uint256 tokenId, address caller);
    error ZeroAddress();
    error EmptyTitle();
    error NotApproved(uint256 tokenId, address caller);

    // ─── Mint ──────────────────────────────────────────────
    /// @notice 在给定 H3 cell mint 一张 GeoyuanCard
    /// @dev    同一 H3 cell 只能 mint 一次 — 地理唯一性
    function mint(
        uint64 h3Index,
        string calldata title,
        string calldata description,
        string calldata color,
        uint16 latE4,
        int16  lonE4,
        uint32 res
    ) external returns (uint256 tokenId) {
        // 唯一性检查：同一 cell 不能 mint 两次
        if (cellMinted[h3Index] != 0) {
            revert CellAlreadyMinted(h3Index, cellMinted[h3Index]);
        }
        if (bytes(title).length == 0) revert EmptyTitle();
        if (msg.sender == address(0)) revert ZeroAddress();

        unchecked { ++totalSupply; }
        tokenId = totalSupply;

        // 核心状态写入
        ownerOf[tokenId]      = msg.sender;
        cellOf[tokenId]       = h3Index;
        cellMinted[h3Index]   = tokenId;
        balanceOf[msg.sender] += 1;
        tokensOfOwner[msg.sender][balanceOf[msg.sender]] = tokenId;

        // Metadata
        metadata[tokenId] = CardMeta({
            title:       title,
            description: description,
            color:       color,
            h3Index:     h3Index,
            latE4:       latE4,
            lonE4:       lonE4,
            res:         res,
            timestamp:   uint40(block.timestamp),
        });
        tokenURI[tokenId] = _buildTokenURI(tokenId);

        emit Transfer(address(0), msg.sender, tokenId);
        emit Minted(tokenId, msg.sender, h3Index);
    }

    // ─── Transfer (简化版 — 纯直转) ────────────────────────
    function transferFrom(address from, address to, uint256 tokenId) external {
        if (ownerOf[tokenId] != from) revert NotOwner(tokenId, from);
        if (to == address(0)) revert ZeroAddress();

        ownerOf[tokenId] = to;
        balanceOf[from]  -= 1;
        balanceOf[to]    += 1;

        emit Transfer(from, to, tokenId);
    }

    // ─── 查询便利函数 ──────────────────────────────────────
    /// @notice 查某个 H3 cell 有没有被 mint
    function isCellMinted(uint64 h3Index) external view returns (bool, uint256) {
        uint256 tid = cellMinted[h3Index];
        return (tid != 0, tid);
    }

    /// @notice 查某地址所有卡片（tokenId 列表）
    function myCards(address who) external view returns (uint256[] memory) {
        uint256 n = balanceOf[who];
        uint256[] memory out = new uint256[](n);
        for (uint256 i = 1; i <= n; i++) {
            out[i - 1] = tokensOfOwner[who][i];
        }
        return out;
    }

    /// @notice 邻居匹配：同 cell 附近 k-ring 内有没有别人的卡
    function neighborMatch(address who, uint64[] calldata neighborCells)
        external view returns (uint256[] memory)
    {
        uint256[] memory out = new uint256[](neighborCells.length);
        uint256 found = 0;
        for (uint256 i = 0; i < neighborCells.length; i++) {
            uint256 tid = cellMinted[neighborCells[i]];
            if (tid != 0 && ownerOf[tid] != who) {
                out[found++] = tid;
            }
        }
        // 截断
        uint256[] memory result = new uint256[](found);
        for (uint256 i = 0; i < found; i++) result[i] = out[i];
        return result;
    }

    // ─── 内部：tokenURI 拼装 ───────────────────────────────
    function _buildTokenURI(uint256 tokenId) internal view returns (string memory) {
        CardMeta storage m = metadata[tokenId];
        // 简化：直接存一个 JSON blob （链上便宜，省掉 IPFS/Arweave 依赖）
        return string(abi.encodePacked(
            '{"name":"Geoyuan #', _itoa(tokenId),
            '","title":"', m.title,
            '","description":"', m.description,
            '","color":"', m.color,
            '","h3Index":"', _itoa(m.h3Index),
            '","lat":', _itoa(int256(m.latE4)),
            ',"lon":', _itoa(int256(m.lonE4)),
            '","res":',  _itoa(int256(m.res)),
            '","timestamp":', _itoa(int256(m.timestamp)),
            '}'
        ));
    }

    // ─── 小工具 ────────────────────────────────────────────
    function _itoa(int256 n) internal pure returns (string memory) {
        if (n == 0) return "0";
        bool neg = n < 0;
        if (neg) n = -n;
        uint256 i;
        uint256 j;
        bytes memory buf = new bytes(78);
        while (n > 0) { buf[i++] = byte('0' + uint8(n % 10)); n /= 10; }
        if (neg) buf[i++] = '-';
        bytes memory out = new bytes(i);
        for (j = 0; j < i; j++) out[j] = buf[i - j - 1];
        return string(out);
    }
}
